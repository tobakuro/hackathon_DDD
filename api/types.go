package main

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
	Memory float64 `json:"memory"`
	CPU    float64 `json:"cpu"`
}

// 4. 攻撃対象コンテナの時系列情報(リザルト用)を渡すためのデータ構造
type AttackHistory struct {
	Timestamp string  `json:"timestamp"`
	Memory    float64 `json:"memory"`
	CPU       float64 `json:"cpu"`
}
