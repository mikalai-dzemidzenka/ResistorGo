package main

import (
	"fmt"
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	bg     = rl.NewColor(5, 12, 10, 255)
	panel  = rl.NewColor(8, 21, 17, 245)
	green  = rl.NewColor(145, 255, 85, 255)
	mint   = rl.NewColor(70, 205, 145, 255)
	white  = rl.NewColor(220, 240, 220, 255)
	muted  = rl.NewColor(75, 115, 90, 255)
	orange = rl.NewColor(255, 155, 65, 255)
)

type Contact struct {
	angle, distance float32
}

func text(s string, x, y, size float32, color rl.Color) {
	rl.DrawTextEx(
		rl.GetFontDefault(), s,
		rl.Vector2{X: x, Y: y},
		size, 2, color,
	)
}

func line(x1, y1, x2, y2, width float32, color rl.Color) {
	rl.DrawLineEx(
		rl.Vector2{X: x1, Y: y1},
		rl.Vector2{X: x2, Y: y2},
		width, color,
	)
}

func polar(cx, cy, radius, angle float32) rl.Vector2 {
	return rl.Vector2{
		X: cx + float32(math.Cos(float64(angle)))*radius,
		Y: cy + float32(math.Sin(float64(angle)))*radius,
	}
}

func frame(r rl.Rectangle, title string) {
	rl.DrawRectangleRec(r, panel)
	rl.DrawRectangleLinesEx(r, 1, rl.Fade(mint, 0.35))

	rl.DrawRectangleRec(rl.Rectangle{
		X: r.X, Y: r.Y, Width: r.Width, Height: 30,
	}, rl.NewColor(15, 36, 26, 255))

	text(title, r.X+12, r.Y+8, 13, green)
	rl.DrawRectangle(int32(r.X+r.Width-19), int32(r.Y+11), 7, 7, green)
}

func drawBackground(w, h, t float32) {
	rl.ClearBackground(bg)

	for x := float32(0); x < w; x += 40 {
		line(x, 0, x, h, 1, rl.Fade(mint, 0.045))
	}
	for y := float32(0); y < h; y += 40 {
		line(0, y, w, y, 1, rl.Fade(mint, 0.045))
	}

	// Движущиеся пакеты данных на краях экрана.
	for i := 0; i < 18; i++ {
		x := float32(math.Mod(
			float64(t*55+float32(i)*113),
			float64(w+100),
		)) - 100

		y := h - 115 + float32(i%3)*13
		line(x, y, x+35, y, 2, rl.Fade(green, 0.12))
	}
}

func drawRadar(cx, cy, radius, t float32, contacts []Contact) {
	tau := float32(2 * math.Pi)
	sweep := float32(math.Mod(float64(t*0.65), float64(tau)))

	rl.DrawCircleV(
		rl.Vector2{X: cx, Y: cy},
		radius, rl.NewColor(7, 25, 17, 255),
	)

	// Затухающий след луча.
	for i := 55; i >= 0; i-- {
		a := sweep - float32(i)*0.009
		p := polar(cx, cy, radius, a)
		alpha := (1 - float32(i)/56) * 0.17
		line(cx, cy, p.X, p.Y, 2, rl.Fade(green, alpha))
	}

	for i := 1; i <= 4; i++ {
		rl.DrawCircleLines(
			int32(cx), int32(cy),
			radius*float32(i)/4,
			rl.Fade(mint, 0.25),
		)
	}

	line(cx-radius, cy, cx+radius, cy, 1, rl.Fade(mint, 0.25))
	line(cx, cy-radius, cx, cy+radius, 1, rl.Fade(mint, 0.25))

	// Деления по окружности.
	for i := 0; i < 60; i++ {
		a := float32(i) * tau / 60
		length := float32(4)
		if i%5 == 0 {
			length = 10
		}
		p1 := polar(cx, cy, radius-length, a)
		p2 := polar(cx, cy, radius, a)
		line(p1.X, p1.Y, p2.X, p2.Y, 1, mint)
	}

	for i, c := range contacts {
		p := polar(cx, cy, radius*c.distance, c.angle)
		age := float32(math.Mod(
			float64(sweep-c.angle+tau), float64(tau),
		))
		alpha := float32(math.Exp(float64(-age * 1.2)))

		rl.DrawCircleV(p, 9, rl.Fade(green, alpha*0.12))
		rl.DrawCircleV(p, 3, rl.Fade(green, 0.08+alpha*0.92))

		if alpha > 0.55 {
			text(fmt.Sprintf("G-%02d", i+1),
				p.X+10, p.Y-8, 12, rl.Fade(green, alpha))
		}
	}

	tip := polar(cx, cy, radius, sweep)
	line(cx, cy, tip.X, tip.Y, 2, rl.Fade(green, 0.9))
	rl.DrawCircleV(rl.Vector2{X: cx, Y: cy}, 4, green)

	text("N", cx-5, cy-radius-20, 13, green)
	text("S", cx-5, cy+radius+9, 13, muted)
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(1280, 760, "GHOST // UPLINK")
	defer rl.CloseWindow()

	rl.SetWindowMinSize(940, 660)
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)

	rng := rand.New(rand.NewSource(2077))
	contacts := make([]Contact, 12)
	for i := range contacts {
		contacts[i] = Contact{
			angle:    rng.Float32() * 2 * math.Pi,
			distance: 0.2 + rng.Float32()*0.67,
		}
	}

	selected := 0
	scanning := true
	inGame := false
	quit := false
	var time float32
	var hoverAmount [4]float32

	for !rl.WindowShouldClose() && !quit {
		dt := rl.GetFrameTime()
		if scanning {
			time += dt
		}

		w := float32(rl.GetScreenWidth())
		h := float32(rl.GetScreenHeight())

		left := rl.Rectangle{
			X: 45, Y: 105,
			Width: 390, Height: h - 180,
		}
		right := rl.Rectangle{
			X: 465, Y: 105,
			Width: w - 510, Height: h - 180,
		}

		items := []string{
			"DEPLOY AGENT",
			"SCANNER: ONLINE",
			"FULLSCREEN",
			"TERMINATE",
		}
		if !scanning {
			items[1] = "SCANNER: PAUSED"
		}

		var buttons [4]rl.Rectangle
		for i := range buttons {
			buttons[i] = rl.Rectangle{
				X:     left.X + 22,
				Y:     left.Y + 205 + float32(i)*53,
				Width: left.Width - 44, Height: 42,
			}
		}

		hovered := -1

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

			activate := rl.IsKeyPressed(rl.KeyEnter) ||
				rl.IsKeyPressed(rl.KeySpace) ||
				(click && hovered >= 0)

			if activate {
				switch selected {
				case 0:
					inGame = true
				case 1:
					scanning = !scanning
					if scanning {
						items[1] = "SCANNER: ONLINE"
					} else {
						items[1] = "SCANNER: PAUSED"
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

		for i := range hoverAmount {
			target := float32(0)
			if i == selected {
				target = 1
			}
			hoverAmount[i] += (target - hoverAmount[i]) *
				float32(1-math.Exp(float64(-dt*16)))
		}

		rl.BeginDrawing()
		drawBackground(w, h, time)

		text("GHOST // UPLINK", 45, 35, 29, green)
		text("REMOTE OPERATIONS DIVISION", 46, 72, 12, muted)
		text("SECURE CHANNEL", w-225, 44, 16, mint)
		line(45, 92, w-45, 92, 1, rl.Fade(mint, 0.35))

		frame(left, "01 / COMMAND")
		frame(right, "02 / PROXIMITY SCANNER")

		text("TRUST", left.X+22, left.Y+53, 44, white)
		text("NO SIGNAL.", left.X+22, left.Y+99, 40, green)
		text("OPERATIVE: GHOST_07", left.X+24, left.Y+157, 14, muted)

		if inGame {
			text("AGENT DEPLOYED", left.X+24, left.Y+225, 25, green)
			text("Awaiting mission data...", left.X+24, left.Y+273, 15, white)
			text("[ ESC ] RETURN", left.X+24, left.Y+345, 17, orange)
		} else {
			for i, r := range buttons {
				a := hoverAmount[i]

				rl.DrawRectangleRec(r, rl.Fade(green, 0.025+a*0.12))
				line(r.X, r.Y+r.Height, r.X+r.Width, r.Y+r.Height,
					1, rl.Fade(mint, 0.15+a*0.55))

				text(fmt.Sprintf("%02d", i+1),
					r.X+10, r.Y+13, 14, muted)
				text(items[i], r.X+47+a*6, r.Y+11, 19, white)
				text("<", r.X+r.Width-25, r.Y+11, 19, rl.Fade(green, a))

				if i == selected {
					rl.DrawRectangle(
						int32(r.X), int32(r.Y+10), 3, 22, green,
					)
				}
			}
		}

		text("ACCESS LEVEL / ROOT", left.X+24, left.Y+left.Height-36, 13, muted)

		// Размер радара зависит от доступного пространства.
		radius := float32(math.Min(
			float64((right.Width-70)/2),
			float64((right.Height-175)/2),
		))
		radarX := right.X + right.Width/2
		radarY := right.Y + 65 + radius

		drawRadar(radarX, radarY, radius, time, contacts)

		infoY := right.Y + right.Height - 80
		line(right.X+20, infoY-14, right.X+right.Width-20, infoY-14,
			1, rl.Fade(mint, 0.3))

		text("CONTACTS / 12", right.X+22, infoY, 15, green)
		text("RANGE / 4.8 KM", right.X+22, infoY+27, 13, muted)

		status := "SWEEP ACTIVE"
		if !scanning {
			status = "SWEEP PAUSED"
		}
		text(status, right.X+right.Width-170, infoY, 13, orange)
		text("ENC / AES-256", right.X+right.Width-170, infoY+27, 13, muted)

		text("W/S  NAVIGATE    ENTER  CONFIRM", 45, h-44, 13, muted)
		text("BUILD 0.9 // OFF-GRID", w-290, h-44, 13, mint)

		rl.EndDrawing()
	}
}
