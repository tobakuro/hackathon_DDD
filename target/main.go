package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

var hitCount int64

func main(){
	http.HandleFunc("/",func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt64(&hitCount, 1)
		log.Printf("ヒット数: %d\n", current)
	})
	
	fmt.Println("標的用サーバーが 8080 で起動しました。")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("死にました: %v\n", err)
	}
}