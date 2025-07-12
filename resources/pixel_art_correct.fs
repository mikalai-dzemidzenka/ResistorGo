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
    
    // Вычисляем производные для определения области интерполяции
    vec2 dudv = fwidth(fragTexCoord);
    
    // Находим ближайший пиксель
    vec2 pixelCoord = floor(fragTexCoord / pixelSize) * pixelSize;
    
    // Вычисляем смещение от центра пикселя
    vec2 offset = fragTexCoord - pixelCoord;
    
    // Нормализуем смещение относительно размера пикселя
    vec2 normalizedOffset = offset / pixelSize;
    
    // Применяем сглаживание только в центре пикселя
    // Используем производные для определения области интерполяции
    vec2 smoothArea = dudv / pixelSize;
    
    // Если мы в области интерполяции, используем сглаживание
    vec2 finalOffset;
    if (normalizedOffset.x < smoothArea.x && normalizedOffset.y < smoothArea.y) {
        // В центре пикселя - используем точные координаты
        finalOffset = vec2(0.5);
    } else {
        // На границе - используем ближайший пиксель
        finalOffset = step(0.5, normalizedOffset);
    }
    
    // Вычисляем финальные координаты
    vec2 finalCoord = pixelCoord + finalOffset * pixelSize;
    
    // Ограничиваем координаты в пределах текстуры
    finalCoord = clamp(finalCoord, vec2(0.0), vec2(1.0));
    
    // Получаем цвет из текстуры
    vec4 texColor = texture(texture0, finalCoord);
    
    // Применяем цвет из vertex shader
    finalColor = texColor * fragColor;
} 