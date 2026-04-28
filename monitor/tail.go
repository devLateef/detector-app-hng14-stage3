package monitor

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"time"

	"detector-app/model"
)

func TailLog(filePath string, out chan<- model.AccessLog) {
	// Retry opening the file until it exists (e.g. nginx hasn't started yet)
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

	// Seek to end so we only process new lines
	file.Seek(0, 2)

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		var entry model.AccessLog
		if err := json.Unmarshal(line, &entry); err == nil {
			out <- entry
		}
	}
}
