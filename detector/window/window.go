// Package window implements deque-based sliding windows for tracking
// per-IP and global request rates over the last 60 seconds.
// Each window entry is a timestamp; eviction removes entries older than 60s.
package window

import (
	"sync"
	"time"
)

const windowDuration = 60 * time.Second

// IPStats tracks per-IP request timestamps and error count within the window
type IPStats struct {
	// Timestamps is a deque of request times within the sliding window
	Timestamps []time.Time
	// Errors counts 4xx/5xx responses within the window
	Errors int
}

// Window holds per-IP and global sliding windows
type Window struct {
	mu     sync.Mutex
	perIP  map[string]*IPStats
	global []time.Time // deque of global request timestamps
}

// NewWindow creates a new Window instance
func NewWindow() *Window {
	return &Window{
		perIP: make(map[string]*IPStats),
	}
}

// Add records a new request from ip with the given HTTP status code.
// It appends to both the per-IP and global deques, then evicts stale entries.
func (w *Window) Add(ip string, status int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()

	if _, ok := w.perIP[ip]; !ok {
		w.perIP[ip] = &IPStats{}
	}

	// Append to per-IP deque
	w.perIP[ip].Timestamps = append(w.perIP[ip].Timestamps, now)

	// Track errors (4xx/5xx)
	if status >= 400 {
		w.perIP[ip].Errors++
	}

	// Append to global deque
	w.global = append(w.global, now)

	// Evict entries older than windowDuration from all deques
	w.evict(now)
}

// evict removes timestamps older than windowDuration from all deques.
// Must be called with w.mu held.
func (w *Window) evict(now time.Time) {
	cutoff := now.Add(-windowDuration)

	// Evict per-IP deques
	for ip, stats := range w.perIP {
		i := 0
		for i < len(stats.Timestamps) && stats.Timestamps[i].Before(cutoff) {
			i++
		}
		stats.Timestamps = stats.Timestamps[i:]

		// Clean up empty entries
		if len(stats.Timestamps) == 0 {
			delete(w.perIP, ip)
		}
	}

	// Evict global deque
	i := 0
	for i < len(w.global) && w.global[i].Before(cutoff) {
		i++
	}
	w.global = w.global[i:]
}

// Rate returns the number of requests from ip in the last 60 seconds
func (w *Window) Rate(ip string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	if stats, ok := w.perIP[ip]; ok {
		return len(stats.Timestamps)
	}
	return 0
}

// ErrorRate returns the number of 4xx/5xx responses from ip in the last 60 seconds
func (w *Window) ErrorRate(ip string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	if stats, ok := w.perIP[ip]; ok {
		return stats.Errors
	}
	return 0
}

// GlobalRate returns the total number of requests across all IPs in the last 60 seconds
func (w *Window) GlobalRate() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.global)
}

// TopIPs returns the top N IPs by request count in the current window
func (w *Window) TopIPs(n int) []IPCount {
	w.mu.Lock()
	defer w.mu.Unlock()

	counts := make([]IPCount, 0, len(w.perIP))
	for ip, stats := range w.perIP {
		counts = append(counts, IPCount{IP: ip, Count: len(stats.Timestamps)})
	}

	// Simple insertion sort (top N is small)
	for i := 1; i < len(counts); i++ {
		for j := i; j > 0 && counts[j].Count > counts[j-1].Count; j-- {
			counts[j], counts[j-1] = counts[j-1], counts[j]
		}
	}

	if n > len(counts) {
		n = len(counts)
	}
	return counts[:n]
}

// IPCount holds an IP and its request count
type IPCount struct {
	IP    string
	Count int
}
