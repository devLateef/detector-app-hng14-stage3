package main

import (
	"detector-app/audit"
	"detector-app/baseline"
	"detector-app/blocker"
	"detector-app/config"
	"detector-app/dashboard"
	"detector-app/detector"
	"detector-app/metrics"
	"detector-app/model"
	"detector-app/monitor"
	"detector-app/notifier"
	"detector-app/unbanner"
	"detector-app/window"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env for local development; in Docker env vars come from docker-compose
	if _, err := os.Stat(".env"); err == nil {
		godotenv.Load()
	}

	config.LoadConfig("config.yaml")

	logChan := make(chan model.AccessLog, 1000)
	win := window.NewWindow()
	base := baseline.NewBaseline()

	// bannedMap tracks currently banned IPs keyed by IP for deduplication.
	// Value is the ban timestamp string.
	var (
		bannedMu  sync.Mutex
		bannedMap = make(map[string]string) // ip -> banned_at
	)

	// Start log tailer
	go monitor.TailLog(config.AppConfig.LogFile, logChan)

	// Start dashboard
	go dashboard.Start()

	// Baseline updater: add global rate sample every second,
	// log BASELINE_RECALC every 60 seconds
	go func() {
		ticker := time.NewTicker(time.Second)
		recalcTicker := time.NewTicker(
			time.Duration(config.AppConfig.Baseline.RecalcIntervalSeconds) * time.Second,
		)
		for {
			select {
			case <-ticker.C:
				globalRate := float64(win.GlobalRate())
				base.Add(globalRate)
				metrics.Set("baseline_mean", base.Mean())
				metrics.Set("baseline_stddev", base.StdDev())

			case <-recalcTicker.C:
				mean := base.Mean()
				std := base.StdDev()
				audit.Log("BASELINE_RECALC", "-", "recalculation", mean, std, 0)
			}
		}
	}()

	// Metrics updater: push dashboard metrics every second
	go func() {
		for {
			time.Sleep(time.Second)
			metrics.Set("global_rate", win.GlobalRate())

			// Top 10 IPs
			top := win.TopIPs(10)
			topList := make([]map[string]any, len(top))
			for i, t := range top {
				topList[i] = map[string]any{"ip": t.IP, "count": t.Count}
			}
			metrics.Set("top_ips", topList)

			// Banned IPs — convert map to list for dashboard
			bannedMu.Lock()
			bannedList := make([]map[string]string, 0, len(bannedMap))
			for ip, bannedAt := range bannedMap {
				bannedList = append(bannedList, map[string]string{
					"ip":        ip,
					"banned_at": bannedAt,
				})
			}
			bannedMu.Unlock()
			metrics.Set("banned_ips", bannedList)

			// Hour slots for baseline graph
			slots := base.HourSlots()
			slotList := make([]map[string]any, 0, len(slots))
			for hour, slot := range slots {
				slotList = append(slotList, map[string]any{
					"hour":   hour,
					"mean":   slot.Mean,
					"stddev": slot.StdDev,
					"count":  slot.Count,
				})
			}
			metrics.Set("hour_slots", slotList)
		}
	}()

	// Main event loop: process each incoming log line
	for log := range logChan {
		win.Add(log.SourceIP, log.Status)

		rate := float64(win.Rate(log.SourceIP))
		globalRate := float64(win.GlobalRate())
		mean := base.Mean()
		std := base.StdDev()
		errorRate := win.ErrorRate(log.SourceIP)
		errorMean := base.ErrorMean()

		// Check for error surge — tightens per-IP detection thresholds
		errorSurge := detector.IsErrorSurge(errorRate, errorMean)

		// Per-IP anomaly detection
		anomaly, reason := detector.IsAnomaly(rate, mean, std, errorSurge)
		if anomaly {
			duration := unbanner.Schedule(log.SourceIP)
			blocker.BlockIP(log.SourceIP)
			audit.Log("BAN", log.SourceIP, reason, rate, mean, duration)

			// Add to banned map (map key deduplicates automatically)
			bannedMu.Lock()
			if _, exists := bannedMap[log.SourceIP]; !exists {
				bannedMap[log.SourceIP] = time.Now().UTC().Format(time.RFC3339)
			}
			bannedMu.Unlock()

			// Schedule removal from banned map after unban
			go func(ip string, dur int) {
				if dur > 0 {
					time.Sleep(time.Duration(dur) * time.Second)
					bannedMu.Lock()
					delete(bannedMap, ip)
					bannedMu.Unlock()
				}
			}(log.SourceIP, duration)

			// Slack ban alert with full context
			banMsg := fmt.Sprintf(
				"*BAN* | IP: `%s` | Condition: %s | Rate: %.2f req/s | Baseline: %.2f | StdDev: %.2f | Duration: %ds | Time: %s",
				log.SourceIP, reason, rate, mean, std, duration,
				time.Now().UTC().Format(time.RFC3339),
			)
			if duration == -1 {
				banMsg = fmt.Sprintf(
					"*PERMANENT BAN* | IP: `%s` | Condition: %s | Rate: %.2f req/s | Baseline: %.2f | Time: %s",
					log.SourceIP, reason, rate, mean,
					time.Now().UTC().Format(time.RFC3339),
				)
			}
			notifier.Send(config.AppConfig.SlackWebhook, banMsg)
		}

		// Global anomaly detection — alert only, no block
		globalAnomaly, globalReason := detector.IsGlobalAnomaly(globalRate, mean, std)
		if globalAnomaly {
			audit.Log("GLOBAL_ALERT", "-", globalReason, globalRate, mean, 0)
			notifier.Send(config.AppConfig.SlackWebhook, fmt.Sprintf(
				"*GLOBAL ANOMALY* | Condition: %s | Global Rate: %.2f req/s | Baseline: %.2f | StdDev: %.2f | Time: %s",
				globalReason, globalRate, mean, std,
				time.Now().UTC().Format(time.RFC3339),
			))
		}
	}
}
