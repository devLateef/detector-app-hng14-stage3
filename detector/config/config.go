package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Thresholds struct {
	ZScore               float64 `yaml:"z_score"`
	Multiplier           float64 `yaml:"multiplier"`
	ErrorSurgeMultiplier float64 `yaml:"error_surge_multiplier"`
}

type BaselineConfig struct {
	WindowSeconds         int     `yaml:"window_seconds"`
	RecalcIntervalSeconds int     `yaml:"recalc_interval_seconds"`
	MinSamples            int     `yaml:"min_samples"`
	FloorMean             float64 `yaml:"floor_mean"`
}

type Config struct {
	LogFile      string         `yaml:"log_file"`
	SlackWebhook string         `yaml:"slack_webhook"`
	Thresholds   Thresholds     `yaml:"thresholds"`
	BanDurations []int          `yaml:"ban_durations"`
	Baseline     BaselineConfig `yaml:"baseline"`
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

	// Override slack webhook from env if set
	if webhook := os.Getenv("SLACK_WEBHOOK"); webhook != "" {
		AppConfig.SlackWebhook = webhook
	}
}
