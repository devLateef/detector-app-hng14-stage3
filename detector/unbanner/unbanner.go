package unbanner

import (
	"fmt"
	"time"

	"detector-app/audit"
	"detector-app/blocker"
	"detector-app/config"
	"detector-app/notifier"
)

var counts = make(map[string]int)

// Schedule sets up an automatic unban for ip after the appropriate backoff duration.
// Returns the ban duration in seconds, or -1 if the ban is permanent.
func Schedule(ip string) int {
	count := counts[ip]

	// Beyond the last duration = permanent ban
	if count >= len(config.AppConfig.BanDurations) {
		return -1
	}

	duration := config.AppConfig.BanDurations[count]
	counts[ip]++

	go func() {
		time.Sleep(time.Duration(duration) * time.Second)
		blocker.UnblockIP(ip)

		// Audit log the unban
		audit.Log("UNBAN", ip, "auto-unban", 0, 0, duration)

		// Slack notification on unban
		notifier.Send(
			config.AppConfig.SlackWebhook,
			fmt.Sprintf(
				"🔓 *UNBAN* | IP: `%s` | Ban duration: %ds | Time: %s",
				ip,
				duration,
				time.Now().UTC().Format(time.RFC3339),
			),
		)
	}()

	return duration
}
