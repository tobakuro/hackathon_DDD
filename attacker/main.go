package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		// 一応デフォルト値(docker-compose.ymlの環境変数で上書きされる)
		targetURL = "http://target-server:8080"
	}

	// 司令塔からの「撃て」を受け取るAPI
	http.HandleFunc("/fire", func(w http.ResponseWriter, r *http.Request) {
		// クエリパラメータ ?count=N を取得（デフォルトは1）
		countStr := r.URL.Query().Get("count")
		count := 1
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 {
			count = c
		}

		log.Printf("命令を受信。標的(%s)に %d 回攻撃します。\n", targetURL, count)

		// 指定回数だけ標的を殴る
		for i := 0; i < count; i++ {
			resp, err := http.Get(targetURL)
			if err != nil {
				log.Printf("攻撃失敗: %v\n", err)
			} else {
				log.Printf("攻撃成功(ステータス: %d)\n", resp.StatusCode)
				defer func() { _ = resp.Body.Close() }()
			}
		}

		_, _ = fmt.Fprintf(w, "%d 回の攻撃、完了しました。\n", count)
	})

	log.Println("攻撃用コンテナ、命令待機中です（ポート80で待機）")
	// コンテナ内で80番ポートを使って司令塔からの通信を待ち受け
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}
