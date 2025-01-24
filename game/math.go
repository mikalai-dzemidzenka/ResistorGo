package game

import (
	"cmp"
	"math"
)

func abs(x float32) float32 {
	return math.Float32frombits(math.Float32bits(x) &^ (1 << 31))
}

func moveTo0(v float32, step float32) float32 {
	switch {
	case v > 0:
		return clamp(v-step, 0, v)
	case v < 0:
		return clamp(v+step, v, 0)
	default:
		return 0
	}
}

func clamp[T cmp.Ordered](v, min, max T) T {
	switch {
	case v > max:
		return max
	case v < min:
		return min
	default:
		return v
	}
}
