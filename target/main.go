package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"hackathon_DDD/target/sound"
)

var hitCount int64
var screamSeq uint64

// in-memory scream buffers (WAV data)
// sound buffers live in package sound

type Config struct {
	ListenAddr      string
	ControlAPIURL   string
	LlamaServerURL  string
	LlamaServerBin  string
	LlamaModelPath  string
	LlamaServerPort string
	LlamaExtraArgs  string
	ScreamPrompt    string
	ScreamMaxTokens int
	ScreamTimeout   time.Duration
	FallbackScream  string
}

type ScreamEvent struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Char    string `json:"char,omitempty"`
	Message string `json:"message,omitempty"`
	Hit     int64  `json:"hit,omitempty"`
}

type ScaleRequest struct {
	Count int `json:"count"`
}

type ScreamGenerator interface {
	Stream(ctx context.Context, prompt string, maxTokens int, onChunk func(string)) error
}

type llamaServerGenerator struct {
	baseURL    string
	httpClient *http.Client
}

type fallbackGenerator struct {
	text string
}

type multiGenerator struct {
	primary  ScreamGenerator
	fallback ScreamGenerator
}

func (g *fallbackGenerator) Stream(_ context.Context, _ string, _ int, onChunk func(string)) error {
	onChunk(g.text)
	return nil
}

func (g *multiGenerator) Stream(ctx context.Context, prompt string, maxTokens int, onChunk func(string)) error {
	if err := g.primary.Stream(ctx, prompt, maxTokens, onChunk); err != nil {
		log.Printf("llama stream failed, using fallback: %v", err)
		return g.fallback.Stream(ctx, prompt, maxTokens, onChunk)
	}
	return nil
}

func (g *llamaServerGenerator) Stream(ctx context.Context, prompt string, maxTokens int, onChunk func(string)) error {
	payload := map[string]any{
		"stream":      true,
		"max_tokens":  maxTokens,
		"temperature": 0.9,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := strings.TrimRight(g.baseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("llama server error: %s: %s", resp.Status, strings.TrimSpace(string(snippet)))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return nil
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				Text string `json:"text"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		content := chunk.Choices[0].Delta.Content
		if content == "" {
			content = chunk.Choices[0].Text
		}
		if content != "" {
			onChunk(content)
		}
	}

	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

type hub struct {
	clients    map[*client]bool
	register   chan *client
	unregister chan *client
	broadcast  chan []byte
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func newHub() *hub {
	return &hub{
		clients:    make(map[*client]bool),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan []byte, 256),
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					delete(h.clients, c)
					close(c.send)
				}
			}
		}
	}
}

func wsHandler(h *hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("websocket upgrade failed: %v", err)
			return
		}
		c := &client{
			conn: conn,
			send: make(chan []byte, 256),
		}
		h.register <- c

		go c.writePump(h)
		c.readPump(h)
	}
}

func (c *client) readPump(h *hub) {
	defer func() {
		h.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *client) writePump(h *hub) {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func broadcastEvent(h *hub, evt ScreamEvent) {
	data, err := json.Marshal(evt)
	if err != nil {
		return
	}
	log.Printf("ws send: %s", data)
	h.broadcast <- data
}

func emitChars(h *hub, id string, chunk string) {
	for _, r := range []rune(chunk) {
		broadcastEvent(h, ScreamEvent{
			ID:   id,
			Type: "char",
			Char: string(r),
		})
	}
}

func hitHandler(h *hub, gen ScreamGenerator, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		current := atomic.AddInt64(&hitCount, 1)
		log.Printf("ヒット数: %d", current)

		id := atomic.AddUint64(&screamSeq, 1)
		idStr := strconv.FormatUint(id, 10)
		broadcastEvent(h, ScreamEvent{ID: idStr, Type: "start", Hit: current})

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hit"))

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.ScreamTimeout)
			defer cancel()
			err := gen.Stream(ctx, cfg.ScreamPrompt, cfg.ScreamMaxTokens, func(chunk string) {
				emitChars(h, idStr, chunk)
			})
			if err != nil {
				broadcastEvent(h, ScreamEvent{ID: idStr, Type: "error", Message: err.Error()})
				return
			}

			sound.SynthesizeAndStore(idStr)
			broadcastEvent(h, ScreamEvent{ID: idStr, Type: "end"})
		}()
	}
}

func scaleProxyHandler(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			http.Error(w, "PUT か POST で送ってください", http.StatusMethodNotAllowed)
			return
		}

		var req ScaleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "JSONが不正です", http.StatusBadRequest)
			return
		}

		if req.Count < 0 {
			http.Error(w, "countは0以上にしてください", http.StatusBadRequest)
			return
		}

		forwardBody, _ := json.Marshal(req)
		targetURL := strings.TrimRight(cfg.ControlAPIURL, "/") + "/scale"
		forwardReq, err := http.NewRequest(http.MethodPut, targetURL, bytes.NewReader(forwardBody))
		if err != nil {
			log.Printf("scale-proxy request create failed: %v", err)
			http.Error(w, "proxy request create failed", http.StatusInternalServerError)
			return
		}
		forwardReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(forwardReq)
		if err != nil {
			log.Printf("scale-proxy forward failed: %v", err)
			http.Error(w, "proxy forward failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		responseBody, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(responseBody)
	}
}

func loadConfig() Config {
	return Config{
		ListenAddr:      getEnv("LISTEN_ADDR", ":8080"),
		ControlAPIURL:   getEnv("CONTROL_API_URL", "http://host.docker.internal:9000"),
		LlamaServerURL:  getEnv("LLAMA_SERVER_URL", "http://127.0.0.1:8081"),
		LlamaServerBin:  os.Getenv("LLAMA_SERVER_BIN"),
		LlamaModelPath:  os.Getenv("LLAMA_MODEL_PATH"),
		LlamaServerPort: getEnv("LLAMA_SERVER_PORT", "8081"),
		LlamaExtraArgs:  os.Getenv("LLAMA_SERVER_ARGS"),
		ScreamPrompt: getEnv(
			"SCREAM_PROMPT",
			"あなたはモンスターです。攻撃を受けた時の短い悲鳴を日本語の擬音で一言だけ出力してください。15文字以内。",
		),
		ScreamMaxTokens: getEnvInt("SCREAM_MAX_TOKENS", 32),
		ScreamTimeout:   getEnvDuration("SCREAM_TIMEOUT", 15*time.Second),
		FallbackScream:  getEnv("FALLBACK_SCREAM", "ギャアアァァ!!"),
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return num
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}

func newGenerator(cfg Config) ScreamGenerator {
	primary := &llamaServerGenerator{
		baseURL:    cfg.LlamaServerURL,
		httpClient: &http.Client{Timeout: cfg.ScreamTimeout},
	}
	fallback := &fallbackGenerator{text: cfg.FallbackScream}
	return &multiGenerator{primary: primary, fallback: fallback}
}

func startLlamaServer(cfg Config) {
	if cfg.LlamaServerBin == "" || cfg.LlamaModelPath == "" {
		return
	}
	if _, err := os.Stat(cfg.LlamaModelPath); err != nil {
		log.Printf("model file not found: %s", cfg.LlamaModelPath)
		return
	}

	args := []string{
		"--model", cfg.LlamaModelPath,
		"--host", "0.0.0.0",
		"--port", cfg.LlamaServerPort,
	}
	if cfg.LlamaExtraArgs != "" {
		args = append(args, strings.Fields(cfg.LlamaExtraArgs)...)
	}

	cmd := exec.Command(cfg.LlamaServerBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("failed to start llama server: %v", err)
		return
	}
	log.Printf("llama server started (pid: %d)", cmd.Process.Pid)
	if err := cmd.Wait(); err != nil {
		log.Printf("llama server exited: %v", err)
	}
}

func main() {
	cfg := loadConfig()
	// read sound config and initialize
	soundSR := getEnvInt("SOUND_SR", 16000)
	soundSeconds := getEnvInt("SOUND_SECONDS", 10)
	soundChannels := getEnvInt("SOUND_CHANNELS", 2)
	soundBitsPerSample := getEnvInt("SOUND_BITS_PER_SAMPLE", 24)
	soundMax := getEnvInt("SOUND_MAX_BUFFERS", 1000)
	sound.Init(soundSR, soundSeconds, soundChannels, soundBitsPerSample, soundMax)
	go startLlamaServer(cfg)

	h := newHub()
	go h.run()
	gen := newGenerator(cfg)

	http.HandleFunc("/", hitHandler(h, gen, cfg))
	http.HandleFunc("/scale-proxy", scaleProxyHandler(cfg))
	http.HandleFunc("/scream/", sound.Handler)
	http.HandleFunc("/ws", wsHandler(h))
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("標的用サーバーが %s で起動しました。", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
		log.Printf("死にました: %v", err)
	}
}
