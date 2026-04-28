package audit

import (
	"fmt"
	"os"
	"time"
)

// Log writes a structured audit entry for ban, unban, and baseline recalculation events.
// Format: [timestamp] ACTION ip | condition | rate | baseline | duration
func Log(action, ip, condition string, rate, baseline float64, duration int) {
	entry := fmt.Sprintf(
		"[%s] %s %s | %s | %.2f | %.2f | %d\n",
		time.Now().UTC().Format(time.RFC3339),
		action,
		ip,
		condition,
		rate,
		baseline,
		duration,
	)

	// Write to stdout so docker logs captures it
	fmt.Print(entry)

	// Persist to audit log file
	f, err := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[audit] failed to open audit.log: %v\n", err)
		return
	}
	defer f.Close()
	f.WriteString(entry)
}
