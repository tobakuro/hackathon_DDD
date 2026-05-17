package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	targetURL := os.Getenv("TARGET_URL")
	if targetURL == "" {
		// 一応デフォルト値(docker-compose.ymlの環境変数で上書きされる)
		targetURL = "http://target-server:8080" 
	}

	fmt.Printf(" %s を攻撃開始\n", targetURL)

	for {
		resp, err := http.Get(targetURL)
		if err != nil {
			fmt.Printf("攻撃失敗: %v\n", err)
		} else {
			fmt.Printf("攻撃成功: %d\n", resp.StatusCode)
			resp.Body.Close()
		}
		// とりあえず100ミリ秒に1回にしておきます
		time.Sleep(100 * time.Millisecond) 
	}
}