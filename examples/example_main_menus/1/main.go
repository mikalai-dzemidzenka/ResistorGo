package main

import (
	"fmt"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	background = rl.NewColor(7, 9, 20, 255)
	panel      = rl.NewColor(12, 17, 34, 240)
	cyan       = rl.NewColor(0, 235, 255, 255)
	pink       = rl.NewColor(255, 40, 145, 255)
	white      = rl.NewColor(225, 240, 255, 255)
	muted      = rl.NewColor(100, 125, 155, 255)
)

type Drop struct {
	x, y, speed, length float32
}

func text(s string, x, y, size float32, color rl.Color) {
	rl.DrawTextEx(
		rl.GetFontDefault(),
		s,
		rl.Vector2{X: x, Y: y},
		size,
		2,
		color,
	)
}

func line(x1, y1, x2, y2, width float32, color rl.Color) {
	rl.DrawLineEx(
		rl.Vector2{X: x1, Y: y1},
		rl.Vector2{X: x2, Y: y2},
		width,
		color,
	)
}

func drawBackground(w, h, t, dt float32, drops []Drop, effects bool) {
	rl.ClearBackground(background)

	rl.DrawRectangleGradientV(
		0, 0, int32(w), int32(h),
		rl.NewColor(15, 10, 32, 255),
		background,
	)

	horizon := h * 0.60

	// Перспективная неоновая сетка.
	for i := -12; i <= 12; i++ {
		bottomX := w/2 + float32(i)*w/10
		line(w/2, horizon, bottomX, h, 1, rl.Fade(cyan, 0.10))
	}

	for i := 0; i < 15; i++ {
		progress := float32(i) / 15
		if effects {
			progress = float32(math.Mod(float64(progress+t*0.07), 1))
		}
		y := horizon + progress*progress*(h-horizon)
		line(0, y, w, y, 1, rl.Fade(cyan, 0.13))
	}

	line(0, horizon, w, horizon, 1, rl.Fade(pink, 0.30))

	if !effects {
		return
	}

	// Дождь: координаты нормализованы для изменения размера окна.
	for i := range drops {
		d := &drops[i]
		d.y += d.speed * dt
		if d.y > 1.15 {
			d.y = -0.15
			d.x = rand.Float32()
		}

		x, y := d.x*w, d.y*h
		line(x, y, x, y+d.length, 1, rl.Fade(cyan, 0.12))
		line(x, y+d.length-3, x, y+d.length, 1, rl.Fade(cyan, 0.45))
	}

	// Сканирующая полоса.
	scanY := float32(math.Mod(float64(t*65), float64(h+100))) - 100
	rl.DrawRectangleGradientV(
		0, int32(scanY), int32(w), 80,
		rl.Fade(cyan, 0),
		rl.Fade(cyan, 0.035),
	)

	// CRT scanlines.
	for y := int32(0); y < int32(h); y += 4 {
		rl.DrawLine(0, y, int32(w), y, rl.Fade(rl.Black, 0.16))
	}
}

func drawButton(rect rl.Rectangle, label string, index int, active bool, t float32) {
	color := muted
	fill := rl.NewColor(16, 23, 42, 235)

	if active {
		color = cyan
		fill = rl.NewColor(15, 45, 60, 245)

		pulse := float32(0.12 + 0.06*math.Sin(float64(t*5)))
		glow := rl.Rectangle{
			X: rect.X - 5, Y: rect.Y - 5,
			Width: rect.Width + 10, Height: rect.Height + 10,
		}
		rl.DrawRectangleRec(glow, rl.Fade(cyan, pulse))
	}

	rl.DrawRectangleRec(rect, fill)
	rl.DrawRectangleLinesEx(rect, 1, rl.Fade(color, 0.65))
	rl.DrawRectangleRec(
		rl.Rectangle{X: rect.X, Y: rect.Y, Width: 4, Height: rect.Height},
		color,
	)

	text(fmt.Sprintf("%02d", index+1), rect.X+18, rect.Y+19, 16, color)
	text(label, rect.X+64, rect.Y+16, 22, white)

	if active {
		text(">", rect.X+rect.Width-30, rect.Y+16, 22, pink)
	}
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 720, "NEON // PROTOCOL")
	defer rl.CloseWindow()

	rl.SetWindowMinSize(800, 600)
	rl.SetTargetFPS(60)
	rl.SetExitKey(0) // Escape обрабатываем самостоятельно.

	drops := make([]Drop, 90)
	for i := range drops {
		drops[i] = Drop{
			x:      rand.Float32(),
			y:      rand.Float32(),
			speed:  0.15 + rand.Float32()*0.35,
			length: 10 + rand.Float32()*35,
		}
	}

	selected := 0
	effects := true
	inGame := false
	quit := false

	for !rl.WindowShouldClose() && !quit {
		dt := rl.GetFrameTime()
		t := float32(rl.GetTime())
		w := float32(rl.GetScreenWidth())
		h := float32(rl.GetScreenHeight())

		menuX := w * 0.09
		menuY := h*0.5 - 55
		menuWidth := float32(390)

		fxLabel := "VISUAL FX: OFF"
		if effects {
			fxLabel = "VISUAL FX: ON"
		}
		labels := []string{
			"START GAME",
			fxLabel,
			"FULLSCREEN",
			"EXIT",
		}

		buttons := make([]rl.Rectangle, len(labels))
		for i := range buttons {
			buttons[i] = rl.Rectangle{
				X: menuX, Y: menuY + float32(i)*68,
				Width: menuWidth, Height: 54,
			}
		}

		if inGame {
			if rl.IsKeyPressed(rl.KeyEscape) {
				inGame = false
			}
		} else {
			if rl.IsKeyPressed(rl.KeyDown) || rl.IsKeyPressed(rl.KeyS) {
				selected = (selected + 1) % len(buttons)
			}
			if rl.IsKeyPressed(rl.KeyUp) || rl.IsKeyPressed(rl.KeyW) {
				selected = (selected + len(buttons) - 1) % len(buttons)
			}

			mouse := rl.GetMousePosition()
			delta := rl.GetMouseDelta()
			click := rl.IsMouseButtonPressed(rl.MouseLeftButton)
			hovered := -1

			for i, rect := range buttons {
				if rl.CheckCollisionPointRec(mouse, rect) {
					hovered = i
					if delta.X != 0 || delta.Y != 0 || click {
						selected = i
					}
				}
			}

			activate := rl.IsKeyPressed(rl.KeyEnter) ||
				rl.IsKeyPressed(rl.KeySpace) ||
				(click && hovered >= 0)

			if activate {
				switch selected {
				case 0:
					inGame = true
				case 1:
					effects = !effects
					if effects {
						labels[1] = "VISUAL FX: ON"
					} else {
						labels[1] = "VISUAL FX: OFF"
					}
				case 2:
					rl.ToggleFullscreen()
				case 3:
					quit = true
				}
			}
		}

		rl.BeginDrawing()
		drawBackground(w, h, t, dt, drops, effects)

		// Верхняя системная панель.
		line(28, 48, w-28, 48, 1, rl.Fade(cyan, 0.35))
		text("N / P", 30, 21, 18, cyan)
		text("SECURE CONNECTION // NIGHT CITY", 130, 23, 14, muted)
		text("SYS.ONLINE", w-150, 23, 14, cyan)

		if inGame {
			text("CONNECTION ESTABLISHED", menuX, h*0.40, 34, cyan)
			text("Your game starts here.", menuX, h*0.40+58, 22, white)
			text("[ ESC ] RETURN TO MENU", menuX, h*0.40+110, 18, pink)
		} else {
			// Полупрозрачная подложка меню.
			rl.DrawRectangleRec(rl.Rectangle{
				X: menuX - 24, Y: menuY - 132,
				Width: menuWidth + 48, Height: 420,
			}, panel)

			text("NEON", menuX+3, menuY-116, 64, rl.Fade(pink, 0.50))
			text("NEON", menuX, menuY-119, 64, cyan)
			text("// PROTOCOL", menuX+3, menuY-52, 25, white)
			text("2077 / ACCESS TERMINAL", menuX, menuY-17, 12, muted)

			for i, rect := range buttons {
				drawButton(rect, labels[i], i, selected == i, t)
			}

			// Декоративный информационный блок справа.
			if w > 1050 {
				x, y := w*0.64, h*0.35
				line(x, y, x+230, y, 2, pink)
				text("DISTRICT_09", x, y+22, 30, white)
				text("STATUS: UNDER SURVEILLANCE", x, y+70, 14, cyan)
				text("IDENTITY: UNKNOWN", x, y+98, 14, muted)
				text("SIGNAL: ENCRYPTED", x, y+126, 14, muted)

				for i := 0; i < 34; i++ {
					height := float32(12)
					if effects {
						height += float32(math.Abs(
							math.Sin(float64(t*2+float32(i)*0.7)),
						)) * 52
					}
					rl.DrawRectangleRec(rl.Rectangle{
						X:     x + float32(i)*7,
						Y:     y + 240 - height,
						Width: 3, Height: height,
					}, rl.Fade(cyan, 0.65))
				}
			}
		}

		line(28, h-48, w-28, h-48, 1, rl.Fade(cyan, 0.25))
		text("W/S  NAVIGATE     ENTER  SELECT", 30, h-31, 14, muted)
		text("BUILD 0.1.7", w-160, h-31, 14, pink)

		rl.EndDrawing()
	}
}
