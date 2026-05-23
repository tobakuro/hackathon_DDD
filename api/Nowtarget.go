package main
import(
    "time"
    "runtime"
    "context"
    "encoding/json"
)
var targetData NowTarget

func streamingTarget() {
    var preSystemCpuUsage, preTargetCpuUsage uint64
    for {
        stats, err := cli.ContainerStats(context.Background(), "target-server", true)
        if err != nil {
            targetData.mu.Lock()
            targetData.IsAlive = false
            targetData.mu.Unlock()
            time.Sleep(1 * time.Second) // エラーが発生した場合は少し待ってから再接続を試みる
            continue
        }
        decoder := json.NewDecoder(stats.Body)// Dockerのストリームからデータを受信するためのデコーダー

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
        stats.Body.Close() // ストリームを閉じる
        time.Sleep(1 * time.Second) // エラーが発生した場合は少し待ってから再接続を試みる
    }
}

func getNowTarget() NowTarget {
    targetData.mu.RLock()
    defer targetData.mu.RUnlock()
    // 読み取った時点のデータのコピーを返す
    return targetData
}