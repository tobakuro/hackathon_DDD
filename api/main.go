package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"bytes"
)

// 1. 増やす命令を受け取るためのデータ構造
type ScaleRequest struct {
	Count int `json:"count"`
}

// 2. 攻撃命令を受け取るためのデータ構造
type AttackRequest struct {
	Target string `json:"target"` // "all" または特定のコンテナなど
	Count  int    `json:"count"`
}

// 3. 今の攻撃対象コンテナの情報を渡すためのデータ構造
type NowTarget struct {
	Memory  float64 `json:"memory"`
	CPU     float64 `json:"cpu"`
}

// 4. 攻撃対象コンテナの時系列情報(リザルト用)を渡すためのデータ構造
type AttackHistory struct {
	Timestamp string `json:"timestamp"`
	Memory    float64    `json:"memory"`
	CPU       float64    `json:"cpu"`
}

// /scale へのリクエストを処理する関数
func scaleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTで送ってください", http.StatusMethodNotAllowed)
		return
	}

	var req ScaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSONの形式が腐っています", http.StatusBadRequest)
		return
	}

	log.Printf("命令を受信。攻撃コンテナを %d 個にスケールします。\n", req.Count)

	// ここで docker compose up --scale attacker=N -d を実行
	cmd := exec.Command("docker", "compose", "up", "--scale", fmt.Sprintf("attacker=%d", req.Count), "-d")
	// 注意: docker-compose.yml があるディレクトリを指定します
	cmd.Dir = "./" 

	if err := cmd.Run(); err != nil {
		log.Printf("スケール失敗: %v\n", err)
		http.Error(w, "Dockerの操作に失敗しました", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "コンテナを %d 個に変更しましたよ\n", req.Count)
}

// /attack へのリクエストを処理する関数
func attackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTで送ってきなさい", http.StatusMethodNotAllowed)
		return
	}

	var req AttackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSONの形式が腐っています", http.StatusBadRequest)
		return
	}

	log.Printf("攻撃命令を受信。対象: %s, 回数: %d\n", req.Target, req.Count)

	// "all" の場合、稼働中のattackerのIPアドレスをDockerコマンドでかき集め
	if req.Target == "all" {
		// すべての attacker コンテナのIPアドレスを取得する
		cmd := exec.Command("docker", "ps", "-q", "--filter", "name=attacker")
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			http.Error(w, "コンテナの取得に失敗", http.StatusInternalServerError)
			return
		}

		containerIDs := strings.Fields(out.String())
		if len(containerIDs) == 0 {
			fmt.Fprintln(w, "attackerが1つも稼働していません")
			return
		}

		// docker exec で直接コンテナ内にコマンドを打ち込み
		for _, id := range containerIDs {
			go func(containerID string, count int) {
				// コンテナの中で、自分自身(127.0.0.1)のAPIをwgetコマンドで叩かせる
				targetUrl := fmt.Sprintf("http://127.0.0.1:80/fire?count=%d", count)
				execCmd := exec.Command("docker", "exec", containerID, "wget", "-qO-", targetUrl)
				
				if err := execCmd.Run(); err != nil {
					log.Printf("attacker(%s)への命令伝達に失敗: %v\n", containerID, err)
				}
			}(id, req.Count)
			
			log.Printf("attacker(%s)に %d 回の攻撃を指示しました\n", id[:8], req.Count)
		 }
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d 体のattackerに攻撃命令を下しました。\n", len(containerIDs))
		
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintf(w, "指定コンテナへの個別命令はまだ未実装\n")
	}
}

// NowTargetへのリクエストを処理する関数
func nowTargetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 一旦ダミーデータを返すだけ
	dummyData := NowTarget{
		Memory: 75.5,
		CPU:    60.3,
	}

	json.NewEncoder(w).Encode(dummyData)
}

// AttckHistoryへのリクエストを処理する関数
func attackHistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 一旦ダミーデータを返すだけ
	dummyHistory := []AttackHistory{
		{Timestamp: "2024-06-01T12:00:00Z", Memory: 70.2, CPU: 55.1},
		{Timestamp: "2024-06-01T12:01:00Z", Memory: 80.5, CPU: 65.3},
		{Timestamp: "2024-06-01T12:02:00Z", Memory: 90.1, CPU: 75.0},
	}

	json.NewEncoder(w).Encode(dummyHistory)
}
func main() {
	http.HandleFunc("/scale", scaleHandler)
	http.HandleFunc("/attack", attackHandler)
	http.HandleFunc("/now-target", nowTargetHandler)
	http.HandleFunc("/attack-history", attackHistoryHandler)

	log.Println("司令塔APIがポート9000で稼働開始")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}