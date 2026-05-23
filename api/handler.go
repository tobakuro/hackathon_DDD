package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"hackathon_DDD/state"
)

func registerRoutes() {
	http.HandleFunc("/scale", scaleHandler)
	http.HandleFunc("/attack", attackHandler)
	http.HandleFunc("/now-target", nowTargetHandler)
	http.HandleFunc("/attack-history", attackHistoryHandler)
	http.HandleFunc("/restart", restartHandler)
}

// /scale へのリクエストを処理する関数
func scaleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "PUTで送ってください", http.StatusMethodNotAllowed)
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
	if r.Method != http.MethodGet {
		http.Error(w, "GETで送ってください", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	data := getNowTarget()
	json.NewEncoder(w).Encode(data)
}

// AttckHistoryへのリクエストを処理する関数
func attackHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GETで送ってください", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	history := getAttackHistory()
	json.NewEncoder(w).Encode(history)
}

// /restart へのリクエストを処理する関数
func restartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POSTで送ってください", http.StatusMethodNotAllowed)
		return
	}

	log.Println("リスタート命令を受信しました")

	// ここで docker compose down && docker compose up -d を実行
	cmd := exec.Command("docker", "compose", "down")
	cmd.Dir = "./"
	if err := cmd.Run(); err != nil {
		log.Printf("リスタート失敗: %v\n", err)
		http.Error(w, "Dockerの操作に失敗しました", http.StatusInternalServerError)
		return
	}

	cmd = exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = "./"
	if err := cmd.Run(); err != nil {
		log.Printf("リスタート失敗: %v\n", err)
		http.Error(w, "Dockerの操作に失敗しました", http.StatusInternalServerError)
		return
	}

	// target-serverコンテナのIDを取得
	containerID, err := getContainerID("target-server")
	if err != nil {
		log.Printf("target-serverコンテナ取得中にエラー: %v\n", err)
		http.Error(w, "Dockerの操作に失敗しました", http.StatusInternalServerError)
		return
	}
	if containerID == "" {
		log.Printf("target-serverコンテナが見つかりませんでした\n")
		http.Error(w, "target-serverコンテナが見つかりませんでした", http.StatusInternalServerError)
		return
	}
	// ステートに保存
	state.SetContainerID(containerID)

	log.Printf("target-serverコンテナを再起動しました (ID: %s)\n", containerID)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RestartResponse{Message: "コンテナを再起動しました"})
}
// リスタートの時に使う関数
func getContainerID(containerName string) (string, error) {

	// フィルターの作成 (name=target-server)
	f := filters.NewArgs()
	// 正確に一致させるために正規表現のアンカーを使用
	f.Add("name", "^/"+containerName+"$")

	// コンテナ一覧の取得
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		Filters: f,
	})
	if err != nil {
		log.Printf("コンテナの取得に失敗しました: %v\n", err)
		return "", err
	}

	if len(containers) == 0 {
		return "", nil // 見つからなかった場合
	}

	return containers[0].ID, nil
}