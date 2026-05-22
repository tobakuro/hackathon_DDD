package main

import (
	"log"
	"net/http"
	"github.com/docker/docker/client"
)

var cli *client.Client

func main() {
	// Dockerクライアントの初期化
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Dockerクライアントの作成に失敗しました: %v", err)
	}
	defer cli.Close()
	registerRoutes()

	log.Println("司令塔APIがポート9000で稼働開始")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
