package metrics

import (
	"runtime"
	"sync"
	"time"
)

var (
	mu        sync.RWMutex
	store     = make(map[string]any)
	startTime = time.Now()
)

// Set stores a key-value metric
func Set(key string, val any) {
	mu.Lock()
	defer mu.Unlock()
	store[key] = val
}

// Get returns a snapshot of all metrics including system stats
func Get() map[string]any {
	mu.RLock()
	defer mu.RUnlock()

	snapshot := make(map[string]any, len(store)+4)
	for k, v := range store {
		snapshot[k] = v
	}

	// Add uptime
	snapshot["uptime_seconds"] = int(time.Since(startTime).Seconds())

	// Add memory usage
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	snapshot["memory_mb"] = float64(mem.Alloc) / 1024 / 1024

	// Add goroutine count as a proxy for CPU activity
	snapshot["goroutines"] = runtime.NumGoroutine()

	return snapshot
}
