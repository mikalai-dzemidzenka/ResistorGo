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

type TimeRewindShader struct {
	Shader       rl.Shader
	TimeLoc      int32
	SpeedLoc     int32
	DirectionLoc int32
	IsEnabled    bool
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

// NewTimeRewindShader создает новый шейдер для визуализации перемотки времени
func NewTimeRewindShader() *TimeRewindShader {
	// Простой vertex shader
	vsCode := `#version 330

in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec4 vertexColor;

out vec2 fragTexCoord;
out vec4 fragColor;

uniform mat4 mvp;

void main() {
    fragTexCoord = vertexTexCoord;
    fragColor = vertexColor;
    gl_Position = mvp * vec4(vertexPosition, 1.0);
}`

	// Fragment shader с эффектами перемотки времени
	fsCode := `#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

out vec4 finalColor;

uniform sampler2D texture0;
uniform float time;
uniform float speed;
uniform float direction;

void main() {
    vec2 uv = fragTexCoord;
    
    // Эффект искажения времени
    float distortion = sin(time * 10.0) * 0.01 * abs(speed);
    uv.x += distortion;
    
    // Цветовой сдвиг в зависимости от направления
    vec4 color = texture(texture0, uv);
    
    if (direction > 0.0) {
        // Перемотка вперед - синий оттенок
        color.rgb = mix(color.rgb, vec3(0.0, 0.5, 1.0), 0.3 * abs(speed));
    } else if (direction < 0.0) {
        // Перемотка назад - красный оттенок
        color.rgb = mix(color.rgb, vec3(1.0, 0.0, 0.0), 0.3 * abs(speed));
    }
    
    // Эффект размытия при высокой скорости
    float blur = abs(speed) * 0.1;
    if (blur > 0.0) {
        vec4 blurColor = vec4(0.0);
        float total = 0.0;
        
        for (int i = -2; i <= 2; i++) {
            for (int j = -2; j <= 2; j++) {
                vec2 offset = vec2(float(i), float(j)) * blur * 0.01;
                blurColor += texture(texture0, uv + offset);
                total += 1.0;
            }
        }
        
        color = mix(color, blurColor / total, min(blur, 0.5));
    }
    
    // Эффект мерцания при перемотке
    float flicker = sin(time * 20.0) * 0.1 + 0.9;
    color.rgb *= flicker;
    
    finalColor = color * fragColor;
}`

	shader := rl.LoadShaderFromMemory(vsCode, fsCode)
	if shader.ID == 0 {
		log.Printf("Ошибка компиляции шейдера перемотки времени")
		return nil
	}

	// Получаем локации uniform переменных
	timeLoc := rl.GetShaderLocation(shader, "time")
	speedLoc := rl.GetShaderLocation(shader, "speed")
	directionLoc := rl.GetShaderLocation(shader, "direction")

	return &TimeRewindShader{
		Shader:       shader,
		TimeLoc:      timeLoc,
		SpeedLoc:     speedLoc,
		DirectionLoc: directionLoc,
		IsEnabled:    true,
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

// BeginTimeRewindMode включает режим визуализации перемотки времени
func (t *TimeRewindShader) BeginTimeRewindMode(time, speed, direction float32) {
	if !t.IsEnabled || t.Shader.ID == 0 {
		return
	}

	// Устанавливаем uniform переменные
	rl.SetShaderValue(t.Shader, t.TimeLoc, []float32{time}, rl.ShaderUniformFloat)
	rl.SetShaderValue(t.Shader, t.SpeedLoc, []float32{speed}, rl.ShaderUniformFloat)
	rl.SetShaderValue(t.Shader, t.DirectionLoc, []float32{direction}, rl.ShaderUniformFloat)

	// Включаем шейдер
	rl.BeginShaderMode(t.Shader)
}

// EndTimeRewindMode выключает режим визуализации перемотки времени
func (t *TimeRewindShader) EndTimeRewindMode() {
	if !t.IsEnabled || t.Shader.ID == 0 {
		return
	}

	rl.EndShaderMode()
}

// Unload освобождает ресурсы шейдера перемотки времени
func (t *TimeRewindShader) Unload() {
	if t.Shader.ID != 0 {
		rl.UnloadShader(t.Shader)
	}
}

// ToggleTimeRewind переключает режим визуализации перемотки времени
func (t *TimeRewindShader) ToggleTimeRewind() {
	t.IsEnabled = !t.IsEnabled
}
