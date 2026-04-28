package notifier

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Send(webhook, message string) {
	payload := map[string]string{"text": message}
	body, _ := json.Marshal(payload)

	http.Post(webhook, "application/json", bytes.NewBuffer(body))
}
