// Package detector implements anomaly detection logic.
// An IP is flagged anomalous if:
//   - Its z-score exceeds the configured threshold (default 3.0), OR
//   - Its rate exceeds multiplier * baseline mean (default 5x)
//
// If an IP has an error surge (4xx/5xx rate >= 3x baseline error mean),
// its detection thresholds are tightened automatically.
package detector

import "detector-app/config"

// IsAnomaly checks whether the given per-IP rate is anomalous relative to
// the rolling baseline mean and stddev. errorSurge tightens thresholds.
// Returns (isAnomaly, reason).
func IsAnomaly(rate, mean, std float64, errorSurge bool) (bool, string) {
	zThreshold := config.AppConfig.Thresholds.ZScore
	multiplier := config.AppConfig.Thresholds.Multiplier

	// Tighten thresholds when the IP has an error surge
	if errorSurge {
		zThreshold *= 0.5
		multiplier *= 0.5
	}

	// Z-score check: how many standard deviations above the mean?
	if std > 0 {
		z := (rate - mean) / std
		if z > zThreshold {
			return true, "z-score"
		}
	}

	// Multiplier check: is the rate more than N times the mean?
	if mean > 0 && rate > multiplier*mean {
		return true, "multiplier"
	}

	return false, ""
}

// IsGlobalAnomaly checks whether the global request rate is anomalous.
// Global anomalies trigger a Slack alert but no IP block.
func IsGlobalAnomaly(globalRate, mean, std float64) (bool, string) {
	return IsAnomaly(globalRate, mean, std, false)
}

// IsErrorSurge returns true if the IP's error rate is >= surgeMultiplier * baseline error mean
func IsErrorSurge(errorRate int, baselineErrorMean float64) bool {
	if baselineErrorMean <= 0 {
		return false
	}
	surgeMultiplier := config.AppConfig.Thresholds.ErrorSurgeMultiplier
	return float64(errorRate) >= surgeMultiplier*baselineErrorMean
}
