package main

import (
	"cmp"
	"math"
)

func abs(x float32) float32 {
	return math.Float32frombits(math.Float32bits(x) &^ (1 << 31))
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
