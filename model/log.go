package model

type AccessLog struct {
	SourceIP     string `json:"source_ip"`
	Timestamp    string `json:"timestamp"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Status       int    `json:"status"`
	ResponseSize int    `json:"response_size"`
}
