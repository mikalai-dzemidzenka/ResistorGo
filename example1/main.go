package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 800
	screenHeight = 450

	// Фиксированный шаг физики/логики (60 обновлений в секунду)
	fixedDt = 1.0 / 60.0

	// Максимальная длина кадра, чтобы избежать "спирали смерти"
	maxFrameTime = 0.25
)

// Actor — простая сущность, которая двигается с постоянной скоростью
type Actor struct {
	// Текущее положение (после последнего фиксированного шага)
	posX, posY float32

	// Положение на предыдущем фиксированном шаге (для интерполяции)
	prevX, prevY float32

	// Скорость (пикселей в секунду)
	velX, velY float32
}

// Update вызывается на каждом фиксированном шаге.
// Здесь происходит изменение состояния объекта.
func (a *Actor) Update() {
	// Сохраняем текущее положение как предыдущее
	a.prevX, a.prevY = a.posX, a.posY

	// Интегрируем скорость (позиция += скорость * dt)
	a.posX += a.velX * fixedDt
	a.posY += a.velY * fixedDt

	// Простая отскок от границ экрана (для наглядности)
	if a.posX < 0 || a.posX > screenWidth {
		a.velX = -a.velX
	}
	if a.posY < 0 || a.posY > screenHeight {
		a.velY = -a.velY
	}
}

// InterpolatedPosition возвращает позицию, интерполированную между
// предыдущим и текущим фиксированным состоянием. alpha ∈ [0,1).
func (a *Actor) InterpolatedPosition(alpha float32) (float32, float32) {
	interpX := a.prevX*(1-alpha) + a.posX*alpha
	interpY := a.prevY*(1-alpha) + a.posY*alpha
	return interpX, interpY
}

func main() {
	rl.InitWindow(screenWidth, screenHeight, "Fixed Timestep Example")
	defer rl.CloseWindow()

	// Ограничиваем FPS (необязательно, но помогает стабильности)
	rl.SetTargetFPS(60)

	// Создаём актора: квадрат, летящий по диагонали
	actor := &Actor{
		posX: 100, posY: 100,
		prevX: 100, prevY: 100, // на старте совпадают
		velX: 200, velY: 150,
	}

	// Аккумулятор времени
	accumulator := 0.0

	for !rl.WindowShouldClose() {
		// Получаем время кадра и ограничиваем его
		frameTime := float64(rl.GetFrameTime())
		if frameTime > maxFrameTime {
			frameTime = maxFrameTime
		}

		// Добавляем время в аккумулятор
		accumulator += frameTime

		// Выполняем все накопившиеся фиксированные шаги
		for accumulator >= fixedDt {
			actor.Update()
			accumulator -= fixedDt
		}

		// Вычисляем коэффициент интерполяции (0..1)
		alpha := float32(accumulator / fixedDt)

		// Получаем интерполированное положение для рендера
		interpX, interpY := actor.InterpolatedPosition(alpha)

		// --- Рендеринг ---
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Рисуем квадрат в интерполированной позиции
		rl.DrawRectangle(int32(interpX), int32(interpY), 50, 50, rl.Maroon)

		rl.EndDrawing()
	}
}
