package main

import (
	"fmt"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	amber  = rl.NewColor(255, 185, 70, 255)
	violet = rl.NewColor(165, 95, 255, 255)
	light  = rl.NewColor(240, 230, 255, 255)
	dim    = rl.NewColor(140, 125, 165, 255)
	dark   = rl.NewColor(9, 7, 19, 255)
)

type Building struct {
	x, width, height float32
	seed             int
}

func label(s string, x, y, size float32, color rl.Color) {
	rl.DrawTextEx(
		rl.GetFontDefault(),
		s,
		rl.Vector2{X: x, Y: y},
		size, 2, color,
	)
}

func centered(s string, cx, y, size float32, color rl.Color) {
	width := rl.MeasureTextEx(rl.GetFontDefault(), s, size, 2).X
	label(s, cx-width/2, y, size, color)
}

func stroke(x1, y1, x2, y2, thickness float32, color rl.Color) {
	rl.DrawLineEx(
		rl.Vector2{X: x1, Y: y1},
		rl.Vector2{X: x2, Y: y2},
		thickness, color,
	)
}

// Угловые скобы вместо обычной рамки.
func brackets(r rl.Rectangle, length float32, color rl.Color) {
	for _, p := range []struct {
		x, y, dx, dy float32
	}{
		{r.X, r.Y, 1, 1},
		{r.X + r.Width, r.Y, -1, 1},
		{r.X, r.Y + r.Height, 1, -1},
		{r.X + r.Width, r.Y + r.Height, -1, -1},
	} {
		stroke(p.x, p.y, p.x+p.dx*length, p.y, 2, color)
		stroke(p.x, p.y, p.x, p.y+p.dy*length, 2, color)
	}
}

func drawCity(w, h, t float32, buildings []Building) {
	rl.ClearBackground(dark)
	rl.DrawRectangleGradientV(
		0, 0, int32(w), int32(h),
		rl.NewColor(30, 12, 54, 255),
		dark,
	)

	// Звёзды с мягким мерцанием.
	for i := 0; i < 85; i++ {
		x := float32((i*193+71)%1000) / 1000 * w
		y := float32((i*137+29)%500) / 1000 * h
		alpha := float32(0.25 + 0.2*math.Sin(float64(t+float32(i))))
		rl.DrawCircleV(rl.Vector2{X: x, Y: y}, 1, rl.Fade(light, alpha))
	}

	// Неоновое солнце.
	sun := rl.Vector2{X: w * 0.76, Y: h * 0.39}
	radius := h * 0.20

	rl.DrawCircleGradient(
		rl.Vector2{sun.X, sun.Y}, radius*1.6,
		rl.Fade(violet, 0.25),
		rl.Fade(violet, 0),
	)
	rl.DrawCircleV(sun, radius, rl.NewColor(210, 77, 130, 255))

	// Горизонтальные разрезы нижней половины солнца.
	for i := 0; i < 10; i++ {
		offset := float32(i) * radius / 10
		halfWidth := float32(math.Sqrt(
			float64(radius*radius - offset*offset),
		))
		stroke(
			sun.X-halfWidth, sun.Y+offset,
			sun.X+halfWidth, sun.Y+offset,
			2+float32(i)*0.6,
			rl.NewColor(29, 12, 47, 255),
		)
	}

	// Силуэты зданий и детерминированный рисунок окон.
	for _, b := range buildings {
		x := b.x * w
		bw := b.width * w
		bh := b.height * h
		y := h - bh

		rl.DrawRectangleRec(rl.Rectangle{
			X: x, Y: y, Width: bw, Height: bh,
		}, rl.NewColor(11, 9, 24, 255))

		stroke(x, y, x+bw, y, 1, rl.Fade(violet, 0.5))

		for row := 0; float32(row)*17+16 < bh; row++ {
			for col := 0; float32(col)*12+10 < bw; col++ {
				if (row*13+col*7+b.seed)%5 < 2 {
					color := rl.Fade(amber, 0.35)
					if (row+col+b.seed)%3 == 0 {
						color = rl.Fade(violet, 0.5)
					}
					rl.DrawRectangle(
						int32(x+6+float32(col)*12),
						int32(y+10+float32(row)*17),
						3, 6, color,
					)
				}
			}
		}

		if b.seed%3 == 0 {
			stroke(x+bw/2, y, x+bw/2, y-25, 1, dim)
			alpha := float32(0.55 + 0.4*math.Sin(float64(t*3)))
			rl.DrawCircleV(
				rl.Vector2{X: x + bw/2, Y: y - 25},
				2, rl.Fade(amber, alpha),
			)
		}
	}

	// CRT-полосы остаются за интерфейсом.
	for y := int32(0); y < int32(h); y += 4 {
		rl.DrawLine(0, y, int32(w), y, rl.Fade(rl.Black, 0.13))
	}
}

func drawReticle(cx, cy, radius, t float32) {
	// Разорванные вращающиеся кольца.
	for ring := 0; ring < 2; ring++ {
		r := radius + float32(ring)*15
		direction := float32(1)
		if ring == 1 {
			direction = -1
		}

		for i := 0; i < 90; i++ {
			if i%30 > 21 {
				continue
			}
			a := float64(float32(i)*2*math.Pi/90 + t*0.15*direction)
			b := a + 2*math.Pi/90

			stroke(
				cx+float32(math.Cos(a))*r,
				cy+float32(math.Sin(a))*r,
				cx+float32(math.Cos(b))*r,
				cy+float32(math.Sin(b))*r,
				1, rl.Fade(violet, 0.3),
			)
		}
	}
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 800, "BLACKOUT // TERMINAL")
	defer rl.CloseWindow()

	rl.SetWindowMinSize(800, 680)
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)

	rng := rand.New(rand.NewSource(42))
	var buildings []Building
	for x := float32(0); x < 1; {
		width := 0.035 + rng.Float32()*0.045
		buildings = append(buildings, Building{
			x:      x,
			width:  width,
			height: 0.16 + rng.Float32()*0.30,
			seed:   rng.Intn(1000),
		})
		x += width + 0.004
	}

	selected := 0
	hovered := -1
	inGame := false
	quit := false
	animated := true
	var clock float32
	var highlight [4]float32

	for !rl.WindowShouldClose() && !quit {
		dt := rl.GetFrameTime()
		if animated {
			clock += dt
		}

		w := float32(rl.GetScreenWidth())
		h := float32(rl.GetScreenHeight())
		cx := w / 2

		card := rl.Rectangle{
			X: cx - 235, Y: h/2 - 245,
			Width: 470, Height: 490,
		}

		items := []string{
			"ENTER THE CITY",
			"ANIMATION: ON",
			"FULLSCREEN",
			"DISCONNECT",
		}
		if !animated {
			items[1] = "ANIMATION: OFF"
		}

		var buttons [4]rl.Rectangle
		for i := range buttons {
			buttons[i] = rl.Rectangle{
				X:     card.X + 35,
				Y:     card.Y + 188 + float32(i)*61,
				Width: card.Width - 70, Height: 48,
			}
		}

		hovered = -1

		if inGame {
			if rl.IsKeyPressed(rl.KeyEscape) {
				inGame = false
			}
		} else {
			if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyS) {
				selected = (selected + 1) % len(items)
			}
			if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyW) {
				selected = (selected + len(items) - 1) % len(items)
			}

			mouse := rl.GetMousePosition()
			delta := rl.GetMouseDelta()
			click := rl.IsMouseButtonPressed(rl.MouseLeftButton)

			for i, r := range buttons {
				if rl.CheckCollisionPointRec(mouse, r) {
					hovered = i
					if delta.X != 0 || delta.Y != 0 || click {
						selected = i
					}
				}
			}

			if rl.IsKeyPressed(rl.KeyEnter) ||
				rl.IsKeyPressed(rl.KeySpace) ||
				(click && hovered >= 0) {
				switch selected {
				case 0:
					inGame = true
				case 1:
					animated = !animated
					if animated {
						items[1] = "ANIMATION: ON"
					} else {
						items[1] = "ANIMATION: OFF"
					}
				case 2:
					rl.ToggleFullscreen()
				case 3:
					quit = true
				}
			}
		}

		if hovered >= 0 && !inGame {
			rl.SetMouseCursor(rl.MouseCursorPointingHand)
		} else {
			rl.SetMouseCursor(rl.MouseCursorDefault)
		}

		// Плавное затухание подсветки при смене выбора.
		for i := range highlight {
			target := float32(0)
			if i == selected {
				target = 1
			}
			factor := float32(1 - math.Exp(float64(-dt*14)))
			highlight[i] += (target - highlight[i]) * factor
		}

		rl.BeginDrawing()
		drawCity(w, h, clock, buildings)

		brackets(rl.Rectangle{
			X: 24, Y: 24, Width: w - 48, Height: h - 48,
		}, 28, rl.Fade(violet, 0.6))

		label("B//T   NETWORK TERMINAL", 45, 43, 14, dim)
		label("SECTOR 07", w-165, 43, 14, amber)

		drawReticle(cx, h/2, 280, clock)

		// Тень и полупрозрачная панель.
		rl.DrawRectangleRec(rl.Rectangle{
			X: card.X + 9, Y: card.Y + 12,
			Width: card.Width, Height: card.Height,
		}, rl.Fade(rl.Black, 0.45))

		rl.DrawRectangleRec(card, rl.NewColor(13, 10, 27, 240))
		rl.DrawRectangleLinesEx(card, 1, rl.Fade(violet, 0.35))
		brackets(card, 22, amber)

		centered("UNAUTHORIZED ACCESS", cx, card.Y+25, 13, amber)

		// Небольшое цветовое смещение заголовка.
		centered("BLACKOUT", cx+2, card.Y+60, 49, rl.Fade(violet, 0.7))
		centered("BLACKOUT", cx, card.Y+58, 49, light)
		centered("// NO SIGNAL. NO RULES.", cx, card.Y+119, 16, dim)

		stroke(
			card.X+35, card.Y+157,
			card.X+card.Width-35, card.Y+157,
			1, rl.Fade(violet, 0.4),
		)

		if inGame {
			centered("IDENTITY ACCEPTED", cx, card.Y+220, 25, amber)
			centered("Welcome to Sector 07.", cx, card.Y+267, 18, light)
			centered("[ ESC ] BACK TO TERMINAL", cx, card.Y+355, 16, dim)
		} else {
			for i, r := range buttons {
				a := highlight[i]

				rl.DrawRectangleRec(r, rl.NewColor(24, 18, 40, 220))
				rl.DrawRectangleRec(r, rl.Fade(amber, a*0.12))
				rl.DrawRectangleLinesEx(r, 1, rl.Fade(amber, a*0.75))

				rl.DrawRectangleRec(rl.Rectangle{
					X: r.X, Y: r.Y,
					Width: 3, Height: r.Height,
				}, rl.Fade(amber, 0.15+a*0.85))

				label(
					fmt.Sprintf("%02d", i+1),
					r.X+15, r.Y+16, 14,
					rl.Fade(amber, 0.4+a*0.6),
				)
				label(items[i], r.X+55+a*5, r.Y+14, 20, light)
				label(">", r.X+r.Width-27, r.Y+14, 20, rl.Fade(amber, a))
			}
		}

		centered("SESSION: GHOST     /     ENCRYPTION: ACTIVE",
			cx, card.Y+457, 12, dim)

		centered("W/S  OR  ARROWS     /     ENTER  TO  CONNECT",
			cx, h-55, 14, dim)

		rl.EndDrawing()
	}
}
