package detector

import "detector-app/config"

func IsAnomaly(rate, mean, std float64) (bool, string) {
	if std > 0 {
		z := (rate - mean) / std
		if z > config.AppConfig.Thresholds.ZScore {
			return true, "z-score"
		}
	}

	if mean > 0 && rate > config.AppConfig.Thresholds.Multiplier*mean {
		return true, "multiplier"
	}

	return false, ""
}
