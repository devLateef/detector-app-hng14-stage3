package detector

func IsAnomaly(rate, mean, std float64) (bool, string) {
	if std > 0 {
		z := (rate - mean) / std
		if z > 3 {
			return true, "z-score"
		}
	}

	if mean > 0 && rate > 5*mean {
		return true, "multiplier"
	}

	return false, ""
}
