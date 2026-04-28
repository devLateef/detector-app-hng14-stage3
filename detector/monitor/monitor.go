package monitor

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"time"

	"detector-app/model"
)

// TailLog continuously tails the nginx access log at filePath,
// parsing each JSON line into an AccessLog and sending it to out.
// It retries opening the file every 3 seconds if it doesn't exist yet.
func TailLog(filePath string, out chan<- model.AccessLog) {
	// Wait for the log file to appear (nginx may not have started yet)
	var file *os.File
	for {
		var err error
		file, err = os.Open(filePath)
		if err == nil {
			break
		}
		log.Printf("[monitor] waiting for log file %s: %v", filePath, err)
		time.Sleep(3 * time.Second)
	}
	defer file.Close()

	// Seek to end — only process new lines, not historical ones
	file.Seek(0, 2)

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			// No new data yet — sleep briefly and retry
			time.Sleep(100 * time.Millisecond)
			continue
		}

		var entry model.AccessLog
		if err := json.Unmarshal(line, &entry); err == nil {
			out <- entry
		}
	}
}
