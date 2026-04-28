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
	"time"

	"github.com/joho/godotenv"
)

func main() {

	godotenv.Load()
	//Load config FIRST
	config.LoadConfig("config.yaml")

	logChan := make(chan model.AccessLog)

	win := window.NewWindow()
	base := baseline.Baseline{}

	// ✅ Use config log path
	go monitor.TailLog(config.AppConfig.LogFile, logChan)

	go dashboard.Start()

	// Baseline updater
	go func() {
		for {
			time.Sleep(time.Second)
			base.Add(float64(win.GlobalRate()))
		}
	}()

	for log := range logChan {
		win.Add(log.SourceIP, log.Status)

		rate := float64(win.Rate(log.SourceIP))
		mean := base.Mean()
		std := base.StdDev()

		metrics.Set("global_rate", win.GlobalRate())

		anomaly, reason := detector.IsAnomaly(rate, mean, std)

		if anomaly {
			duration := unbanner.Schedule(log.SourceIP)

			blocker.BlockIP(log.SourceIP)

			audit.Log("BAN", log.SourceIP, reason, rate, mean, duration)

			// Use config webhook
			notifier.Send(
				os.Getenv("SLACK_WEBHOOK"),
				fmt.Sprintf("🚨 Blocked %s (%s) | rate=%.2f baseline=%.2f",
					log.SourceIP, reason, rate, mean),
			)
		}
	}
}
