package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Thresholds struct {
	ZScore     float64 `yaml:"z_score"`
	Multiplier float64 `yaml:"multiplier"`
}

type Config struct {
	LogFile      string     `yaml:"log_file"`
	SlackWebhook string     `yaml:"slack_webhook"`
	Thresholds   Thresholds `yaml:"thresholds"`
	BanDurations []int      `yaml:"ban_durations"`
}

var AppConfig Config

func LoadConfig(path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	err = yaml.Unmarshal(file, &AppConfig)
	if err != nil {
		panic("Failed to parse config file: " + err.Error())
	}
}
