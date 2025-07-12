package gamepkg

import (
	"io/ioutil"
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PixelArtShader struct {
	Shader    rl.Shader
	ZoomLoc   int32
	ResLoc    int32
	IsEnabled bool
	Type      int // 0 - basic, 1 - simple, 2 - advanced
}

// NewPixelArtShader создает новый pixel art шейдер
func NewPixelArtShader() *PixelArtShader {
	return NewPixelArtShaderWithType(0) // По умолчанию используем базовый шейдер
}

// NewPixelArtShaderWithType создает pixel art шейдер указанного типа
func NewPixelArtShaderWithType(shaderType int) *PixelArtShader {
	// Загружаем vertex shader
	vsCode, err := ioutil.ReadFile("resources/pixel_art.vs")
	if err != nil {
		log.Printf("Ошибка загрузки vertex shader: %v", err)
		return nil
	}

	// Выбираем fragment shader в зависимости от типа
	var fsPath string
	switch shaderType {
	case 0:
		fsPath = "resources/pixel_art.fs"
	case 1:
		fsPath = "resources/pixel_art_simple.fs"
	case 2:
		fsPath = "resources/pixel_art_advanced.fs"
	case 3:
		fsPath = "resources/pixel_art_integer.fs"
	default:
		fsPath = "resources/pixel_art.fs"
	}

	fsCode, err := ioutil.ReadFile(fsPath)
	if err != nil {
		log.Printf("Ошибка загрузки fragment shader (%s): %v", fsPath, err)
		return nil
	}

	shader := rl.LoadShaderFromMemory(string(vsCode), string(fsCode))
	if shader.ID == 0 {
		log.Printf("Ошибка компиляции шейдера типа %d", shaderType)
		return nil
	}

	// Получаем локации uniform переменных
	zoomLoc := rl.GetShaderLocation(shader, "zoom")
	resLoc := rl.GetShaderLocation(shader, "resolution")

	return &PixelArtShader{
		Shader:    shader,
		ZoomLoc:   zoomLoc,
		ResLoc:    resLoc,
		IsEnabled: true,
		Type:      shaderType,
	}
}

// BeginPixelArtMode включает pixel art режим
func (p *PixelArtShader) BeginPixelArtMode(camera *rl.Camera2D) {
	if !p.IsEnabled || p.Shader.ID == 0 {
		return
	}

	// Устанавливаем uniform переменные
	rl.SetShaderValue(p.Shader, p.ZoomLoc, []float32{camera.Zoom}, rl.ShaderUniformFloat)
	rl.SetShaderValue(p.Shader, p.ResLoc, []float32{float32(ScreenWidth), float32(ScreenHeight)}, rl.ShaderUniformVec2)

	// Включаем шейдер
	rl.BeginShaderMode(p.Shader)
}

// EndPixelArtMode выключает pixel art режим
func (p *PixelArtShader) EndPixelArtMode() {
	if !p.IsEnabled || p.Shader.ID == 0 {
		return
	}

	rl.EndShaderMode()
}

// Unload освобождает ресурсы шейдера
func (p *PixelArtShader) Unload() {
	if p.Shader.ID != 0 {
		rl.UnloadShader(p.Shader)
	}
}

// TogglePixelArt переключает pixel art режим
func (p *PixelArtShader) TogglePixelArt() {
	p.IsEnabled = !p.IsEnabled
}

// SwitchShaderType переключает тип шейдера
func (p *PixelArtShader) SwitchShaderType() {
	// Освобождаем текущий шейдер
	if p.Shader.ID != 0 {
		rl.UnloadShader(p.Shader)
	}

	// Переключаемся на следующий тип
	p.Type = (p.Type + 1) % 4

	// Создаем новый шейдер
	vsCode, err := ioutil.ReadFile("resources/pixel_art.vs")
	if err != nil {
		log.Printf("Ошибка загрузки vertex shader: %v", err)
		return
	}

	var fsPath string
	switch p.Type {
	case 0:
		fsPath = "resources/pixel_art.fs"
	case 1:
		fsPath = "resources/pixel_art_simple.fs"
	case 2:
		fsPath = "resources/pixel_art_advanced.fs"
	case 3:
		fsPath = "resources/pixel_art_integer.fs"
	}

	fsCode, err := ioutil.ReadFile(fsPath)
	if err != nil {
		log.Printf("Ошибка загрузки fragment shader (%s): %v", fsPath, err)
		return
	}

	shader := rl.LoadShaderFromMemory(string(vsCode), string(fsCode))
	if shader.ID == 0 {
		log.Printf("Ошибка компиляции шейдера типа %d", p.Type)
		return
	}

	p.Shader = shader
	p.ZoomLoc = rl.GetShaderLocation(shader, "zoom")
	p.ResLoc = rl.GetShaderLocation(shader, "resolution")
}

// GetShaderTypeName возвращает название типа шейдера
func (p *PixelArtShader) GetShaderTypeName() string {
	switch p.Type {
	case 0:
		return "Basic"
	case 1:
		return "Simple"
	case 2:
		return "Advanced"
	case 3:
		return "Integer"
	default:
		return "Unknown"
	}
}
