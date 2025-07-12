#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

uniform sampler2D texture0;
uniform vec2 resolution;
uniform float zoom;

out vec4 finalColor;

// Функция для вычисления модифицированных UV координат
vec2 pixelArtUV(vec2 uv, vec2 texSize) {
    vec2 pixel = uv * texSize;
    
    // Находим ближайший шов между текселями
    vec2 seam = floor(pixel + 0.5);
    
    // Вычисляем производные для определения области интерполяции
    vec2 dudv = fwidth(pixel);
    
    // Модифицируем координаты пикселя относительно шва
    pixel = seam + clamp((pixel - seam) / dudv, -0.5, 0.5);
    
    return pixel / texSize;
}

void main() {
    // Получаем размер текстуры
    vec2 texSize = textureSize(texture0, 0);
    
    // Применяем pixel art фильтрацию
    vec2 pixelCoord = pixelArtUV(fragTexCoord, texSize);
    
    // Ограничиваем координаты в пределах текстуры
    pixelCoord = clamp(pixelCoord, vec2(0.0), vec2(1.0));
    
    // Получаем цвет из текстуры
    vec4 texColor = texture(texture0, pixelCoord);
    
    // Применяем цвет из vertex shader
    finalColor = texColor * fragColor;
} 