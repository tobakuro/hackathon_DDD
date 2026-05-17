package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
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
	cmd.Dir = "../" 

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
	// 今はただ返事をするだけのハリボテ
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "まだ攻撃の実装してない\n")
}

func main() {
	http.HandleFunc("/scale", scaleHandler)
	http.HandleFunc("/attack", attackHandler)

	log.Println("司令塔APIがポート9000で稼働開始")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}