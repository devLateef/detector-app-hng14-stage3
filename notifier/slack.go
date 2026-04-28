package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Send(webhook, message string) {
	if webhook == "" {
		fmt.Println("[notifier] SLACK_WEBHOOK is empty, skipping notification")
		return
	}

	payload := map[string]string{"text": message}
	body, _ := json.Marshal(payload)

	fmt.Printf("[notifier] sending to Slack: %s\n", message)

	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("[notifier] error sending to Slack: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("[notifier] Slack response: %d %s\n", resp.StatusCode, string(respBody))
}
