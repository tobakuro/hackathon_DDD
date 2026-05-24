package main

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"time"
)

var targetData NowTarget

// watchIsAlive はIsAliveがfalseになったタイミングでPodの自動再起動を発行する
func watchIsAlive() {
	prevAlive := true
	for {
		time.Sleep(1 * time.Second)
		current := getNowTarget()
		if prevAlive && !current.IsAlive {
			log.Println("IsAlive=falseを検知。target-serverを自動再起動します")
			if err := restartTargetServer(); err != nil {
				log.Printf("自動再起動失敗: %v\n", err)
			}
		}
		prevAlive = current.IsAlive
	}
}

func streamingTarget() {
	for {
		var preSystemCpuUsage, preTargetCpuUsage uint64
		stats, err := cli.ContainerStats(context.Background(), "target-server", true)
		if err != nil {
			targetData.mu.Lock()
			targetData.IsAlive = false
			targetData.mu.Unlock()
			time.Sleep(1 * time.Second) // エラーが発生した場合は少し待ってから再接続を試みる
			continue
		}
		decoder := json.NewDecoder(stats.Body) // Dockerのストリームからデータを受信するためのデコーダー

		for {
			// 1. データ受信 (待ち時間発生、ロックしない)
			var stats DockerStats
			if err := decoder.Decode(&stats); err != nil {
				targetData.mu.Lock()
				targetData.IsAlive = false
				targetData.mu.Unlock()
				break
			}

			if preSystemCpuUsage != 0 {
				newSystemCpuUsage := stats.CpuStats.SystemCpuUsage
				newTargetCpuUsage := stats.CpuStats.CpuUsage.TotalUsage
				// 2. CPU使用率の計算 (ロックしない)
				// new-pre<0の場合は初期化
				if newSystemCpuUsage < preSystemCpuUsage || newTargetCpuUsage < preTargetCpuUsage {
					preSystemCpuUsage = 0
					preTargetCpuUsage = 0
					continue
				}
				newCPU := (float64(newTargetCpuUsage-preTargetCpuUsage) / (float64(newSystemCpuUsage-preSystemCpuUsage) * 0.05)) * float64(runtime.NumCPU()) * 100.0
				// 3. メモリ使用率の計算 (ロックしない)
				newMemory := float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
				// 4. 安全に書き込み (一瞬だけロックする)
				targetData.mu.Lock()
				targetData.IsAlive = true
				targetData.CPU = newCPU
				targetData.Memory = newMemory
				targetData.mu.Unlock()
			} else {
				targetData.mu.Lock()
				targetData.IsAlive = true
				targetData.mu.Unlock()
			}
			// 4. 次のループの計算のために履歴を更新
			preSystemCpuUsage = stats.CpuStats.SystemCpuUsage
			preTargetCpuUsage = stats.CpuStats.CpuUsage.TotalUsage
		}
		stats.Body.Close()          // ストリームを閉じる
		time.Sleep(1 * time.Second) // エラーが発生した場合は少し待ってから再接続を試みる
	}
}

func getNowTarget() NowTarget {
	targetData.mu.RLock()
	defer targetData.mu.RUnlock()
	// mutex 自体は返さず、JSON 化に必要な値だけをコピーして返す
	return NowTarget{
		Memory:  targetData.Memory,
		CPU:     targetData.CPU,
		IsAlive: targetData.IsAlive,
	}
}
