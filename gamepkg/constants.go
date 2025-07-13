package gamepkg

const (
	ScreenWidth         = 800
	ScreenHeight        = 450
	SpriteSize          = 48
	SpriteScale         = 1.0
	Gravity             = 0.0
	JumpSpeed           = 0.0
	AttackFps           = 24
	AttackFrameTime     = 0.45 / 8               // 0.45 секунды на 8 кадров (атака)
	BackAttackFrameTime = float32(4) / AttackFps // 4 кадра для возврата (~0.1667 сек)
	HitFrameTime        = 0.3333 / 4             // 0.08333 сек/кадр (4 кадра, 12 FPS, 0.3333 сек)
	RunFps              = 18
	RunFrameTime        = 1.0 / RunFps
	AttackBoxWidth      = 24 * SpriteScale // 24 пикселя
	AttackBoxHeight     = 6 * SpriteScale  // 6 пикселей
	MinZoom             = 0.5              // Минимальный масштаб камеры
	MaxZoom             = 8.0              // Максимальный масштаб камеры
	ZoomSpeed           = 0.1              // Скорость изменения масштаба
	MoveAcceleration    = 0.95             // Ускорение движения (было 0.9)
	MoveDeceleration    = 0.98             // Замедление движения (было 0.95)
	MaxSpeed            = 50.0             // Максимальная скорость (ускорено)

	UpDownTransitionDuration  = 0.1
	UpDownTransitionFrameTime = UpDownTransitionDuration / 4
)
