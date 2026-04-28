package window

import (
	"sync"
	"time"
)

type Stats struct {
	Timestamps []time.Time
	Errors     int
}

type Window struct {
	mu     sync.Mutex
	perIP  map[string]*Stats
	global []time.Time
}

func NewWindow() *Window {
	return &Window{
		perIP: make(map[string]*Stats),
	}
}

func (w *Window) Add(ip string, status int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()

	if _, ok := w.perIP[ip]; !ok {
		w.perIP[ip] = &Stats{}
	}

	w.perIP[ip].Timestamps = append(w.perIP[ip].Timestamps, now)
	w.global = append(w.global, now)

	if status >= 400 {
		w.perIP[ip].Errors++
	}

	w.evict()
}

func (w *Window) evict() {
	cutoff := time.Now().Add(-60 * time.Second)

	for _, stat := range w.perIP {
		filtered := []time.Time{}
		for _, t := range stat.Timestamps {
			if t.After(cutoff) {
				filtered = append(filtered, t)
			}
		}
		stat.Timestamps = filtered
	}

	filteredGlobal := []time.Time{}
	for _, t := range w.global {
		if t.After(cutoff) {
			filteredGlobal = append(filteredGlobal, t)
		}
	}
	w.global = filteredGlobal
}

func (w *Window) Rate(ip string) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.perIP[ip].Timestamps)
}

func (w *Window) GlobalRate() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.global)
}
