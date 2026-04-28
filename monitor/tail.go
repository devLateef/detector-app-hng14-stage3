package monitor

import (
	"bufio"
	"encoding/json"
	"os"

	"detector-app/model"
)

func TailLog(filePath string, out chan<- model.AccessLog) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	file.Seek(0, 2)

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			continue
		}

		var log model.AccessLog
		if err := json.Unmarshal(line, &log); err == nil {
			out <- log
		}
	}
}
