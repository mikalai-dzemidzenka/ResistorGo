package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	canvasW = 480
	canvasH = 270
)

var (
	sky   = rl.NewColor(10, 5, 19, 255)
	red   = rl.NewColor(255, 45, 72, 255)
	hot   = rl.NewColor(255, 125, 130, 255)
	pale  = rl.NewColor(255, 225, 225, 255)
	muted = rl.NewColor(155, 100, 120, 255)
)

type Star struct {
	x, y  int32
	phase float32
	large bool
}

type Vec3 struct {
	x, y, z float32
}

func text(s string, x, y, size float32, color rl.Color) {
	rl.DrawTextEx(
		rl.GetFontDefault(),
		s,
		rl.Vector2{X: x, Y: y},
		size,
		1,
		color,
	)
}

func centered(s string, y, size float32, color rl.Color) {
	width := rl.MeasureTextEx(rl.GetFontDefault(), s, size, 1).X
	x := float32(math.Floor(float64((canvasW - width) / 2)))
	text(s, x, y, size, color)
}

func screenDestination() (rl.Rectangle, float32) {
	scale := float32(math.Floor(math.Min(
		float64(rl.GetScreenWidth())/canvasW,
		float64(rl.GetScreenHeight())/canvasH,
	)))
	if scale < 1 {
		scale = 1
	}

	return rl.Rectangle{
		X:      float32(math.Floor(float64((float32(rl.GetScreenWidth()) - canvasW*scale) / 2))),
		Y:      float32(math.Floor(float64((float32(rl.GetScreenHeight()) - canvasH*scale) / 2))),
		Width:  canvasW * scale,
		Height: canvasH * scale,
	}, scale
}

// Планета из горизонтальных полос с пиксельным краем.
func pixelPlanet(cx, cy, radius int32) {
	for dy := -radius; dy < radius; dy += 2 {
		half := int32(math.Sqrt(float64(radius*radius - dy*dy)))
		if half == 0 {
			continue
		}

		color := rl.NewColor(118, 29, 55, 255)
		if dy < -radius/3 {
			color = rl.NewColor(188, 57, 73, 255)
		} else if dy > radius/3 {
			color = rl.NewColor(67, 18, 44, 255)
		}

		if (dy+radius)/2%7 == 0 {
			color = rl.NewColor(139, 37, 64, 255)
		}

		rl.DrawRectangle(cx-half, cy+dy, half*2, 2, color)

		// Теневая сторона.
		start := half / 3
		rl.DrawRectangle(
			cx+start, cy+dy, half-start, 2,
			rl.NewColor(35, 12, 34, 255),
		)
	}
}

func landscape(t float32, stars []Star) {
	rl.ClearBackground(sky)

	// Ступенчатый градиент неба.
	for i := int32(0); i < 18; i++ {
		rl.DrawRectangle(
			0, i*10, canvasW, 10,
			rl.NewColor(
				uint8(10+i),
				5,
				uint8(19+i/3),
				255,
			),
		)
	}

	for _, s := range stars {
		alpha := float32(0.55 + 0.35*math.Sin(float64(t*1.4+s.phase)))
		color := rl.Fade(pale, alpha)

		rl.DrawRectangle(s.x, s.y, 1, 1, color)

		if s.large {
			rl.DrawRectangle(s.x-1, s.y, 3, 1, rl.Fade(color, 0.6))
			rl.DrawRectangle(s.x, s.y-1, 1, 3, rl.Fade(color, 0.6))
		}
	}

	pixelPlanet(383, 67, 32)

	// Далёкий спутник.
	rl.DrawRectangle(81, 55, 6, 6, muted)
	rl.DrawRectangle(81, 55, 4, 4, pale)

	// Два слоя гор.
	for layer := 0; layer < 2; layer++ {
		for x := int32(0); x < canvasW; x += 4 {
			f := float64(x)
			height := 18 +
				math.Sin(f*0.024+float64(layer)*2)*13 +
				math.Sin(f*0.065+float64(layer))*7

			y := int32(151-height) + int32(layer)*15

			color := rl.NewColor(49, 17, 41, 255)
			if layer == 1 {
				color = rl.NewColor(23, 10, 27, 255)
			}

			rl.DrawRectangle(x, y, 4, 175-y, color)
		}
	}

	const horizon int32 = 170

	rl.DrawRectangle(
		0, horizon, canvasW, canvasH-horizon,
		rl.NewColor(17, 5, 17, 255),
	)

	// Свечение горизонта.
	for i := int32(8); i >= 1; i-- {
		rl.DrawRectangle(
			0, horizon-i, canvasW, i*2,
			rl.Fade(red, 0.025),
		)
	}
	rl.DrawLine(0, horizon, canvasW, horizon, hot)

	// Перспективная сетка.
	for i := -12; i <= 12; i++ {
		endX := int32(canvasW/2 + i*50)
		rl.DrawLine(
			canvasW/2, horizon,
			endX, canvasH,
			rl.Fade(red, 0.32),
		)
	}

	// Поперечные линии движутся к зрителю.
	for i := 0; i < 12; i++ {
		p := float32(math.Mod(float64(float32(i)/12+t*0.065), 1))
		y := horizon + int32(p*p*float32(canvasH-horizon))

		rl.DrawRectangle(0, y-1, canvasW, 3, rl.Fade(red, 0.06))
		rl.DrawLine(0, y, canvasW, y, rl.Fade(red, 0.2+p*0.5))
	}

	// Неоновая пиксельная трава.
	for i := 0; i < 85; i++ {
		x := int32((i*71 + 13) % canvasW)
		y := int32(180 + (i*37)%90)

		if x > 145 && x < 335 {
			continue
		}

		height := int32(2 + i%6)
		color := rl.NewColor(71, 17, 38, 255)
		if i%4 == 0 {
			color = red
		}

		rl.DrawRectangle(x, y-height, 1, height, color)
		rl.DrawRectangle(x-1, y-height+1, 1, 2, color)
	}

	// Поднимающиеся искры.
	for i := 0; i < 22; i++ {
		x := int32((i*97 + 21) % canvasW)
		rise := float32(math.Mod(float64(t*8+float32(i)*11), 85))
		y := int32(239 - rise)

		rl.DrawRectangle(x, y, 1, 2, rl.Fade(hot, 0.5))
	}
}

func drawAI(cx, cy, t float32) {
	phi := float32((1 + math.Sqrt(5)) / 2)

	// Внешняя оболочка — икосаэдр.
	vertices := []Vec3{
		{-1, phi, 0}, {1, phi, 0},
		{-1, -phi, 0}, {1, -phi, 0},
		{0, -1, phi}, {0, 1, phi},
		{0, -1, -phi}, {0, 1, -phi},
		{phi, 0, -1}, {phi, 0, 1},
		{-phi, 0, -1}, {-phi, 0, 1},
	}

	rotationY := t * 0.38
	rotationX := float32(0.35) + t*0.19

	sy := float32(math.Sin(float64(rotationY)))
	cyRot := float32(math.Cos(float64(rotationY)))
	sx := float32(math.Sin(float64(rotationX)))
	cxRot := float32(math.Cos(float64(rotationX)))

	// Вращение в 3D и перспективная проекция в пиксельный экран.
	project := func(v Vec3, size float32) (rl.Vector2, float32) {
		x := v.x*cyRot + v.z*sy
		z := -v.x*sy + v.z*cyRot

		y := v.y*cxRot - z*sx
		z = v.y*sx + z*cxRot

		x *= size
		y *= size
		z *= size

		perspective := float32(220) / (220 - z)

		return rl.Vector2{
			X: float32(math.Round(float64(cx + x*perspective))),
			Y: float32(math.Round(float64(cy + y*perspective))),
		}, z
	}

	drawEdge := func(a, b rl.Vector2, alpha float32) {
		rl.DrawLineEx(a, b, 3, rl.Fade(red, alpha*0.09))
		rl.DrawLine(
			int32(a.X), int32(a.Y),
			int32(b.X), int32(b.Y),
			rl.Fade(red, alpha),
		)
	}

	// Разорванная орбита.
	for i := 0; i < 80; i++ {
		if i%20 > 13 {
			continue
		}

		a := float64(i)*2*math.Pi/80 + float64(t)*0.12
		b := a + 2*math.Pi/80

		x1 := cx + float32(math.Cos(a))*68
		y1 := cy + float32(math.Sin(a))*20
		x2 := cx + float32(math.Cos(b))*68
		y2 := cy + float32(math.Sin(b))*20

		rl.DrawLine(
			int32(x1), int32(y1),
			int32(x2), int32(y2),
			rl.Fade(red, 0.22),
		)
	}

	size := float32(25) + float32(math.Sin(float64(t*1.6)))*0.8

	var points [12]rl.Vector2
	var depths [12]float32

	for i, v := range vertices {
		points[i], depths[i] = project(v, size)
	}

	// Задние рёбра темнее передних.
	for pass := 0; pass < 2; pass++ {
		for i := 0; i < len(vertices); i++ {
			for j := i + 1; j < len(vertices); j++ {
				dx := vertices[i].x - vertices[j].x
				dy := vertices[i].y - vertices[j].y
				dz := vertices[i].z - vertices[j].z

				// Длина ребра икосаэдра равна 2.
				if math.Abs(float64(dx*dx+dy*dy+dz*dz-4)) > 0.01 {
					continue
				}

				depth := (depths[i] + depths[j]) / 2
				front := depth >= 0

				if (pass == 0 && front) || (pass == 1 && !front) {
					continue
				}

				alpha := float32(0.28)
				if front {
					alpha = 0.85
				}

				drawEdge(points[i], points[j], alpha)
			}
		}
	}

	// Узлы оболочки.
	for i, p := range points {
		color := rl.Fade(red, 0.45)
		if depths[i] >= 0 {
			color = hot
		}

		rl.DrawRectangle(int32(p.X)-1, int32(p.Y)-1, 2, 2, color)
	}

	// Внутренний каркас — октаэдр.
	core := []Vec3{
		{1, 0, 0}, {-1, 0, 0},
		{0, 1, 0}, {0, -1, 0},
		{0, 0, 1}, {0, 0, -1},
	}

	for i := 0; i < len(core); i++ {
		for j := i + 1; j < len(core); j++ {
			if i/2 == j/2 {
				continue
			}

			a, _ := project(core[i], 14)
			b, _ := project(core[j], 14)
			drawEdge(a, b, 0.7)
		}
	}

	// Пульсирующее сердце ИИ.
	pulse := float32(0.65 + 0.35*math.Sin(float64(t*3)))

	rl.DrawRectangle(
		int32(cx)-7, int32(cy)-7,
		14, 14, rl.Fade(red, pulse*0.08),
	)
	rl.DrawRectangle(
		int32(cx)-3, int32(cy)-3,
		6, 6, rl.Fade(red, pulse*0.6),
	)
	rl.DrawRectangle(int32(cx)-1, int32(cy)-1, 2, 2, pale)

	// Пакеты данных.
	for i := 0; i < 3; i++ {
		a := float64(t*0.9) + float64(i)*2*math.Pi/3
		x := cx + float32(math.Cos(a))*68
		y := cy + float32(math.Sin(a))*20

		rl.DrawRectangle(int32(x)-1, int32(y)-1, 3, 3, hot)
	}

	// Прицельные скобы.
	for _, side := range []float32{-1, 1} {
		x := int32(cx + side*59)
		innerX := x - int32(side*7)

		color := rl.Fade(red, 0.4)
		rl.DrawLine(x, int32(cy)-10, x, int32(cy)+10, color)
		rl.DrawLine(x, int32(cy)-10, innerX, int32(cy)-10, color)
		rl.DrawLine(x, int32(cy)+10, innerX, int32(cy)+10, color)
	}
}

func drawButton(r rl.Rectangle, title string, active bool, t float32) {
	x, y := int32(r.X), int32(r.Y)
	w, h := int32(r.Width), int32(r.Height)

	rl.DrawRectangle(x+2, y+3, w, h, rl.NewColor(5, 3, 10, 230))

	border := rl.NewColor(92, 34, 54, 255)
	fill := rl.NewColor(29, 12, 27, 255)
	color := muted

	if active {
		border = red
		fill = rl.NewColor(64, 17, 36, 255)
		color = pale

		rl.DrawRectangle(x-2, y-2, w+4, h+4, rl.Fade(red, 0.12))
	}

	// Пиксельные срезанные углы.
	rl.DrawRectangle(x+1, y, w-2, h, border)
	rl.DrawRectangle(x, y+1, w, h-2, border)
	rl.DrawRectangle(x+2, y+2, w-4, h-4, fill)

	rl.DrawLine(x+2, y+2, x+w-3, y+2, rl.Fade(hot, 0.2))

	width := rl.MeasureTextEx(rl.GetFontDefault(), title, 10, 1).X
	text(
		title,
		float32(math.Floor(float64(r.X+(r.Width-width)/2))),
		r.Y+6,
		10,
		color,
	)

	if active {
		offset := int32(0)
		if math.Sin(float64(t*5)) > 0 {
			offset = 2
		}

		text(">", float32(x-12-offset), float32(y+6), 10, hot)
		text("<", float32(x+w+6+offset), float32(y+6), 10, hot)
	}
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(1440, 810, "RESISTOR // ECHO")
	defer rl.CloseWindow()

	rl.SetWindowMinSize(960, 540)
	rl.SetTargetFPS(60)
	rl.SetExitKey(0)

	// Низкое внутреннее разрешение для пиксельного стиля.
	target := rl.LoadRenderTexture(canvasW, canvasH)
	defer rl.UnloadRenderTexture(target)
	rl.SetTextureFilter(target.Texture, rl.FilterPoint)

	rng := rand.New(rand.NewSource(77))
	stars := make([]Star, 95)

	for i := range stars {
		stars[i] = Star{
			x:     int32(rng.Intn(canvasW)),
			y:     int32(rng.Intn(140)),
			phase: rng.Float32() * 2 * math.Pi,
			large: i%13 == 0,
		}
	}

	selected := 0
	animated := true
	inGame := false
	quit := false
	var clock float32

	for !rl.WindowShouldClose() && !quit {
		if animated {
			clock += rl.GetFrameTime()
		}

		destination, scale := screenDestination()

		mouse := rl.GetMousePosition()
		virtualMouse := rl.Vector2{
			X: (mouse.X - destination.X) / scale,
			Y: (mouse.Y - destination.Y) / scale,
		}

		items := []string{
			"CONNECT",
			"AMBIENCE: ON",
			"FULLSCREEN",
			"DISCONNECT",
		}
		if !animated {
			items[1] = "AMBIENCE: OFF"
		}

		// Меню слева снизу.
		var buttons [4]rl.Rectangle
		for i := range buttons {
			buttons[i] = rl.Rectangle{
				X:      24,
				Y:      137 + float32(i)*25,
				Width:  140,
				Height: 22,
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

			delta := rl.GetMouseDelta()
			click := rl.IsMouseButtonPressed(rl.MouseLeftButton)

			for i, r := range buttons {
				if rl.CheckCollisionPointRec(virtualMouse, r) {
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
					animated = !animated
					if animated {
						items[1] = "AMBIENCE: ON"
					} else {
						items[1] = "AMBIENCE: OFF"
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

		// Обновляем область вывода после переключения fullscreen.
		destination, _ = screenDestination()

		rl.BeginTextureMode(target)
		landscape(clock, stars)

		// Объёмный пиксельный заголовок.
		centered("RESISTOR", 34, 30, rl.Fade(red, 0.2))
		centered("RESISTOR", 30, 30, rl.Fade(red, 0.3))
		centered("RESISTOR", 33, 30, rl.NewColor(90, 15, 38, 255))
		centered("RESISTOR", 31, 30, pale)

		// Световая проекция ИИ на поле.
		for i := int32(0); i < 4; i++ {
			halfWidth := int32(24 - i*5)
			rl.DrawRectangle(
				canvasW/2-halfWidth,
				204+i*2,
				halfWidth*2,
				1,
				rl.Fade(red, 0.16-float32(i)*0.03),
			)
		}

		// Парящее ядро по центру.
		aiY := float32(144) +
			float32(math.Sin(float64(clock*1.2)))*3

		drawAI(canvasW/2, aiY, clock)
		centered("ECHO / AI CORE", 221, 10, hot)

		text("LINK", 397, 207, 10, muted)

		status := "ACTIVE"
		if !animated {
			status = "PAUSED"
		}
		text(status, 397, 221, 10, hot)

		if inGame {
			// Здесь можно подключить собственную игровую сцену.
			rl.DrawRectangle(
				18, 135, 148, 98,
				rl.NewColor(16, 7, 22, 235),
			)
			rl.DrawRectangle(18, 135, 2, 98, red)

			text("LINK ESTABLISHED", 28, 149, 10, pale)
			text("ECHO IS LISTENING", 28, 169, 10, muted)
			text("[ ESC ] RETURN", 28, 207, 10, hot)
		} else {
			text("EXPLORE THE UNKNOWN", 24, 118, 10, muted)
			rl.DrawLine(24, 130, 164, 130, rl.Fade(red, 0.35))

			for i, r := range buttons {
				drawButton(r, items[i], i == selected, clock)
			}
		}

		text("W/S : SELECT   ENTER : CONFIRM", 24, 251, 10, muted)
		rl.EndTextureMode()

		// Увеличение без размытия.
		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(4, 2, 8, 255))

		rl.DrawTexturePro(
			target.Texture,
			rl.Rectangle{
				Width:  canvasW,
				Height: -canvasH,
			},
			destination,
			rl.Vector2{},
			0,
			rl.White,
		)

		rl.EndDrawing()
	}
}
