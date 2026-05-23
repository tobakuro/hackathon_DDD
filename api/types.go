package main

import "sync"

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
	IsAlive bool    `json:"is_alive"`
	mu      sync.RWMutex
}

type DockerStats struct {
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	CpuStats struct {
		CpuUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCpuUsage uint64 `json:"system_cpu_usage"`
	} `json:"cpu_stats"`
}

// 4. 攻撃対象コンテナの時系列情報(リザルト用)を渡すためのデータ構造
type AttackHistory struct {
	Timestamp string  `json:"timestamp"`
	Memory    float64 `json:"memory"`
	CPU       float64 `json:"cpu"`
}

// 5. リスタート命令を受け取った後の完了レスポンスのデータ構造
type RestartResponse struct {
	Message string `json:"message"`
}
