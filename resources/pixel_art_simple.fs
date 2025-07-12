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
    
    // Находим центр текущего пикселя
    vec2 pixelCenter = floor(fragTexCoord / pixelSize) * pixelSize + pixelSize * 0.5;
    
    // Вычисляем расстояние от текущей точки до центра пикселя
    vec2 offset = fragTexCoord - pixelCenter;
    
    // Если мы находимся близко к границе пикселя, используем ближайший пиксель
    vec2 threshold = pixelSize * 0.1; // Порог для определения границы
    
    vec2 finalCoord;
    if (abs(offset.x) < threshold.x && abs(offset.y) < threshold.y) {
        // Мы в центре пикселя - используем точные координаты
        finalCoord = pixelCenter;
    } else {
        // Мы на границе - используем ближайший пиксель
        finalCoord = floor(fragTexCoord / pixelSize) * pixelSize + pixelSize * 0.5;
    }
    
    // Ограничиваем координаты в пределах текстуры
    finalCoord = clamp(finalCoord, vec2(0.0), vec2(1.0));
    
    // Получаем цвет из текстуры
    vec4 texColor = texture(texture0, finalCoord);
    
    // Применяем цвет из vertex shader
    finalColor = texColor * fragColor;
} 