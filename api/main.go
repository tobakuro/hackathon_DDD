package main

import (
	"log"
	"net/http"
)

func main() {
	registerRoutes()

	log.Println("司令塔APIがポート9000で稼働開始")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
