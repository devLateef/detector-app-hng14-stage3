package unbanner

import (
	"time"

	"detector-app/blocker"
	"detector-app/config"
)

var counts = make(map[string]int)

func Schedule(ip string) int {
	count := counts[ip]

	if count >= len(config.AppConfig.BanDurations) {
		return -1
	}

	duration := config.AppConfig.BanDurations[count]
	counts[ip]++

	go func() {
		time.Sleep(time.Duration(duration) * time.Second)
		blocker.UnblockIP(ip)
	}()

	return duration
}
