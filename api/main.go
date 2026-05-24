package main

import (
	"github.com/docker/docker/client"
	"log"
	"net/http"
)

var cli *client.Client

func main() {
	// Dockerクライアントの初期化
	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Dockerクライアントの作成に失敗しました: %v", err)
	}
	defer func() { _ = cli.Close() }()

	// ターゲットのリソース使用率をストリーミングで取得するゴルーチンを開始
	go streamingTarget()
	// IsAliveの変化を監視して自動再起動を発行するゴルーチンを開始
	go watchIsAlive()

	// HTTPサーバーのルートを登録
	registerRoutes()

	log.Println("司令塔APIがポート9000で稼働開始")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
