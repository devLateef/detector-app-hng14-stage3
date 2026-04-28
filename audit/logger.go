package audit

import (
	"fmt"
	"os"
	"time"
)

func Log(action, ip, condition string, rate, baseline float64, duration int) {
	entry := fmt.Sprintf(
		"[%s] %s %s | %s | %.2f | %.2f | %d\n",
		time.Now().Format(time.RFC3339),
		action,
		ip,
		condition,
		rate,
		baseline,
		duration,
	)

	// Write to stdout so docker logs captures it
	fmt.Print(entry)

	// Also persist to file
	f, _ := os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(entry)
}
