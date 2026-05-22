package state

import "sync"

var (
	mu          sync.RWMutex
	containerID string
)

// SetContainerID はスレッドセーフにIDを保存
func SetContainerID(id string) {
	mu.Lock()
	defer mu.Unlock()
	containerID = id
}

// GetContainerID はスレッドセーフにIDを取得
func GetContainerID() string {
	mu.RLock()
	defer mu.RUnlock()
	return containerID
}