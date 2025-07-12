#version 330

in vec2 fragTexCoord;
in vec4 fragColor;

uniform sampler2D texture0;
uniform vec2 resolution;
uniform float zoom;

out vec4 finalColor;

void main() {
    // Получаем размер текстуры
    vec2 texSize = textureSize(texture0, 0);
    
    // Вычисляем размер пикселя в текстуре
    vec2 pixelSize = 1.0 / texSize;
    
    // Вычисляем целочисленный масштаб
    float integerZoom = floor(zoom);
    float fractionalZoom = zoom - integerZoom;
    
    // Если масштаб близок к целому числу, используем целочисленное масштабирование
    vec2 finalCoord;
    if (fractionalZoom < 0.1) {
        // Используем целочисленный масштаб
        vec2 scaledCoord = floor(fragTexCoord * integerZoom) / integerZoom;
        vec2 pixelCoord = floor(scaledCoord / pixelSize) * pixelSize + pixelSize * 0.5;
        finalCoord = clamp(pixelCoord, vec2(0.0), vec2(1.0));
    } else {
        // Используем обычное pixel art масштабирование
        vec2 pixelCoord = floor(fragTexCoord / pixelSize) * pixelSize + pixelSize * 0.5;
        finalCoord = clamp(pixelCoord, vec2(0.0), vec2(1.0));
    }
    
    // Получаем цвет из текстуры
    vec4 texColor = texture(texture0, finalCoord);
    
    // Применяем цвет из vertex shader
    finalColor = texColor * fragColor;
} 