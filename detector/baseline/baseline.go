// Package baseline implements a rolling baseline calculator.
// It maintains a 30-minute window of per-second request counts,
// recalculates mean/stddev every 60 seconds, and keeps per-hour
// slots to prefer the current hour's data when it has enough samples.
package baseline

import (
	"math"
	"sync"
	"time"

	"detector-app/config"
)

// sample holds a per-second count with its timestamp
type sample struct {
	ts    time.Time
	count float64
}

// HourSlot holds aggregated baseline stats for a single hour
type HourSlot struct {
	Mean   float64
	StdDev float64
	Count  int
}

// Baseline tracks rolling request rate statistics
type Baseline struct {
	mu sync.Mutex

	// Rolling window of per-second samples (last 30 minutes)
	samples []sample

	// Cached mean and stddev, recalculated every 60 seconds
	effectiveMean   float64
	effectiveStdDev float64
	lastRecalc      time.Time

	// Per-hour slots: key = hour (0-23)
	hourSlots map[int]*HourSlot

	// Error rate baseline (for error surge detection)
	errorSamples []sample
	errorMean    float64
}

// NewBaseline creates a new Baseline instance
func NewBaseline() *Baseline {
	return &Baseline{
		hourSlots: make(map[int]*HourSlot),
	}
}

// Add records a new per-second global request count sample
func (b *Baseline) Add(count float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	windowDur := time.Duration(config.AppConfig.Baseline.WindowSeconds) * time.Second

	// Append new sample
	b.samples = append(b.samples, sample{ts: now, count: count})

	// Evict samples older than the rolling window
	cutoff := now.Add(-windowDur)
	start := 0
	for start < len(b.samples) && b.samples[start].ts.Before(cutoff) {
		start++
	}
	b.samples = b.samples[start:]

	// Recalculate every RecalcIntervalSeconds
	recalcInterval := time.Duration(config.AppConfig.Baseline.RecalcIntervalSeconds) * time.Second
	if now.Sub(b.lastRecalc) >= recalcInterval {
		b.recalculate(now)
		b.lastRecalc = now
	}
}

// AddError records a per-second error count for error surge detection
func (b *Baseline) AddError(count float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	windowDur := time.Duration(config.AppConfig.Baseline.WindowSeconds) * time.Second

	b.errorSamples = append(b.errorSamples, sample{ts: now, count: count})

	cutoff := now.Add(-windowDur)
	start := 0
	for start < len(b.errorSamples) && b.errorSamples[start].ts.Before(cutoff) {
		start++
	}
	b.errorSamples = b.errorSamples[start:]

	// Recalculate error mean
	if len(b.errorSamples) > 0 {
		sum := 0.0
		for _, s := range b.errorSamples {
			sum += s.count
		}
		b.errorMean = sum / float64(len(b.errorSamples))
	}
}

// recalculate computes mean and stddev from current samples and updates hour slots.
// Must be called with b.mu held.
func (b *Baseline) recalculate(now time.Time) {
	n := len(b.samples)
	if n == 0 {
		return
	}

	// Compute mean
	sum := 0.0
	for _, s := range b.samples {
		sum += s.count
	}
	mean := sum / float64(n)

	// Compute stddev
	variance := 0.0
	for _, s := range b.samples {
		diff := s.count - mean
		variance += diff * diff
	}
	stddev := math.Sqrt(variance / float64(n))

	// Apply floor to mean
	floor := config.AppConfig.Baseline.FloorMean
	if mean < floor {
		mean = floor
	}

	b.effectiveMean = mean
	b.effectiveStdDev = stddev

	// Update per-hour slot
	hour := now.Hour()
	if _, ok := b.hourSlots[hour]; !ok {
		b.hourSlots[hour] = &HourSlot{}
	}
	slot := b.hourSlots[hour]
	slot.Mean = mean
	slot.StdDev = stddev
	slot.Count = n
}

// Mean returns the effective mean, preferring current hour's slot if it has enough data
func (b *Baseline) Mean() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	hour := time.Now().Hour()
	if slot, ok := b.hourSlots[hour]; ok && slot.Count >= config.AppConfig.Baseline.MinSamples {
		return slot.Mean
	}

	if b.effectiveMean < config.AppConfig.Baseline.FloorMean {
		return config.AppConfig.Baseline.FloorMean
	}
	return b.effectiveMean
}

// StdDev returns the effective standard deviation
func (b *Baseline) StdDev() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	hour := time.Now().Hour()
	if slot, ok := b.hourSlots[hour]; ok && slot.Count >= config.AppConfig.Baseline.MinSamples {
		return slot.StdDev
	}
	return b.effectiveStdDev
}

// ErrorMean returns the baseline error rate mean
func (b *Baseline) ErrorMean() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.errorMean
}

// HourSlots returns a copy of all hour slots for dashboard display
func (b *Baseline) HourSlots() map[int]HourSlot {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := make(map[int]HourSlot, len(b.hourSlots))
	for k, v := range b.hourSlots {
		result[k] = *v
	}
	return result
}

// SampleCount returns the number of samples in the rolling window
func (b *Baseline) SampleCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.samples)
}
