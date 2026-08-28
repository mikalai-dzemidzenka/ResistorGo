package resistor

import (
	"cmp"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

var ZeroVec2 = rl.Vector2{}

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

func sign(x float32) float32 { // -1, 0 or 1
	if x > -0.00001 || x < 0.00001 {
		return 0
	}
	return (float32(math.Float32bits(x)&(1<<31)) - 0.5) * 2
}

func lerp[T float32 | float64](a, b, t T) T {
	return a + t*(b-a)
}

func clamp[T cmp.Ordered](v, minimum, maximum T) T {
	return max(minimum, min(v, maximum))
}

func abs(x float32) float32 {
	return math.Float32frombits(math.Float32bits(x) &^ (1 << 31))
}

func aabb(a, b rl.Rectangle) bool {
	return a.X < b.X+b.Width && a.X+a.Width > b.X && a.Y < b.Y+b.Height && a.Y+a.Height > b.Y
}

func distance(a, b rl.Vector2) float32 {
	return length(sub(a, b))
}

func length(v rl.Vector2) float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

func sub(a, b rl.Vector2) rl.Vector2 {
	return rl.Vector2{
		X: a.X - b.X,
		Y: a.Y - b.Y,
	}
}
