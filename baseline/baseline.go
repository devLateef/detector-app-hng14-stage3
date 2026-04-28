package baseline

import "math"

type Baseline struct {
	values []float64
}

func (b *Baseline) Add(v float64) {
	b.values = append(b.values, v)
	if len(b.values) > 1800 {
		b.values = b.values[1:]
	}
}

func (b *Baseline) Mean() float64 {
	if len(b.values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range b.values {
		sum += v
	}
	return sum / float64(len(b.values))
}

func (b *Baseline) StdDev() float64 {
	if len(b.values) == 0 {
		return 0
	}

	mean := b.Mean()
	sum := 0.0

	for _, v := range b.values {
		sum += math.Pow(v-mean, 2)
	}

	return math.Sqrt(sum / float64(len(b.values)))
}
