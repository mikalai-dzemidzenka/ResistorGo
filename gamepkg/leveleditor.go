package gamepkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type LevelElementType int

const (
	BlockElement LevelElementType = iota
	EnemyElement
)

type Block struct {
	Position rl.Vector2
	Size     rl.Vector2
}

// EnemyData упрощенная структура для сохранения врагов в JSON
type EnemyData struct {
	Position rl.Vector2
	Type     ActorType
}

type Level struct {
	Blocks  []Block
	Enemies []EnemyData
}

type LevelEditor struct {
	ActiveElementType LevelElementType
	ShowDropdown      bool
	DropdownActive    int32
	Level             Level
	DisplayEnemies    []Actor // Враги для отображения в редакторе
}

func NewLevelEditor() LevelEditor {
	return LevelEditor{
		ActiveElementType: BlockElement,
		ShowDropdown:      false,
		DropdownActive:    0,
		Level: Level{
			Blocks:  []Block{},
			Enemies: []EnemyData{},
		},
	}
}

func (le *LevelEditor) SaveLevel(filename string) error {
	fmt.Printf("Saving level: %d blocks, %d enemies\n", len(le.Level.Blocks), len(le.Level.Enemies))
	data, err := json.MarshalIndent(le.Level, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling level: %v\n", err)
		return err
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return err
	}
	fmt.Printf("Level saved successfully to %s\n", filename)
	return nil
}

func (le *LevelEditor) LoadLevel(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File %s not found, creating empty level\n", filename)
		// Если файла нет, инициализируем пустой уровень
		le.Level = Level{
			Blocks:  []Block{},
			Enemies: []EnemyData{},
		}
		return nil
	}
	err = json.Unmarshal(data, &le.Level)
	if err != nil {
		fmt.Printf("Error unmarshaling level: %v\n", err)
		return err
	}
	fmt.Printf("Loaded level: %d blocks, %d enemies\n", len(le.Level.Blocks), len(le.Level.Enemies))
	return nil
}

func (le *LevelEditor) Update(camera *rl.Camera2D, scale float32, posX, posY float32) {
	// Переключение типа элемента через Tab
	if rl.IsKeyPressed(rl.KeyTab) {
		le.ShowDropdown = !le.ShowDropdown
	}

	// Обработка UI (выпадающий список) - используем фиксированные экранные координаты
	if le.ShowDropdown {
		items := []string{"Block", "Enemy"}
		uiPosX := float32(10) // Фиксированная позиция
		uiPosY := float32(10) // Фиксированная позиция
		dropdownRect := rl.NewRectangle(uiPosX, uiPosY+60, 120, 30)
		if raygui.DropdownBox(dropdownRect, stringsJoin(items, ";"), &le.DropdownActive, le.ShowDropdown) {
			le.ActiveElementType = LevelElementType(le.DropdownActive)
			le.ShowDropdown = false
		}
	}

	// Размещение элемента (ЛКМ)
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !le.ShowDropdown {
		// Преобразуем координаты мыши с учетом масштабирования
		mousePos := rl.GetMousePosition()

		// Проверяем, находится ли мышь в пределах масштабированного изображения
		if mousePos.X >= posX && mousePos.X <= posX+float32(ScreenWidth)*scale &&
			mousePos.Y >= posY && mousePos.Y <= posY+float32(ScreenHeight)*scale {

			// Нормализуем координаты мыши относительно масштабированного изображения
			normalizedX := (mousePos.X - posX) / (float32(ScreenWidth) * scale)
			normalizedY := (mousePos.Y - posY) / (float32(ScreenHeight) * scale)

			// Преобразуем в координаты сцены (без учета камеры)
			sceneX := normalizedX * float32(ScreenWidth)
			sceneY := normalizedY * float32(ScreenHeight)

			// Применяем обратное преобразование камеры
			worldPos := rl.GetScreenToWorld2D(rl.NewVector2(sceneX, sceneY), *camera)

			switch le.ActiveElementType {
			case BlockElement:
				block := Block{
					Position: rl.NewVector2(float32(int(worldPos.X/48)*48), float32(int(worldPos.Y/48)*48)), // Привязка к сетке 48x48
					Size:     rl.NewVector2(48*SpriteScale, 48*SpriteScale),
				}
				le.Level.Blocks = append(le.Level.Blocks, block)
			case EnemyElement:
				enemyData := EnemyData{
					Position: rl.NewVector2(worldPos.X, ScreenHeight-90*SpriteScale),
					Type:     AIActor,
				}
				le.Level.Enemies = append(le.Level.Enemies, enemyData)
			}
		}
	}

	// Удаление элемента (ПКМ)
	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		// Преобразуем координаты мыши с учетом масштабирования
		mousePos := rl.GetMousePosition()

		// Проверяем, находится ли мышь в пределах масштабированного изображения
		if mousePos.X >= posX && mousePos.X <= posX+float32(ScreenWidth)*scale &&
			mousePos.Y >= posY && mousePos.Y <= posY+float32(ScreenHeight)*scale {

			// Нормализуем координаты мыши относительно масштабированного изображения
			normalizedX := (mousePos.X - posX) / (float32(ScreenWidth) * scale)
			normalizedY := (mousePos.Y - posY) / (float32(ScreenHeight) * scale)

			// Преобразуем в координаты сцены (без учета камеры)
			sceneX := normalizedX * float32(ScreenWidth)
			sceneY := normalizedY * float32(ScreenHeight)

			// Применяем обратное преобразование камеры
			worldPos := rl.GetScreenToWorld2D(rl.NewVector2(sceneX, sceneY), *camera)

			// Удаление блоков
			newBlocks := []Block{}
			for _, block := range le.Level.Blocks {
				blockRect := rl.NewRectangle(block.Position.X, block.Position.Y, block.Size.X, block.Size.Y)
				if !rl.CheckCollisionPointRec(worldPos, blockRect) {
					newBlocks = append(newBlocks, block)
				}
			}
			le.Level.Blocks = newBlocks

			// Удаление врагов
			newEnemies := []EnemyData{}
			for _, enemy := range le.Level.Enemies {
				enemyRect := rl.NewRectangle(enemy.Position.X, enemy.Position.Y, 48*SpriteScale, 48*SpriteScale)
				if !rl.CheckCollisionPointRec(worldPos, enemyRect) {
					newEnemies = append(newEnemies, enemy)
				}
			}
			le.Level.Enemies = newEnemies
		}
	}
}

// DrawScene рисует только сцену редактора (без UI)
func (le *LevelEditor) DrawScene(camera *rl.Camera2D) {
	// Отрисовка блоков
	for _, block := range le.Level.Blocks {
		rl.DrawRectangleV(block.Position, block.Size, rl.Gray)
	}

	// Отрисовка врагов
	for _, enemyData := range le.Level.Enemies {
		// Создаем временного актора для отрисовки
		enemy := NewActor(enemyData.Position, enemyData.Type)
		DrawActor(&enemy, nil, rl.Blue, camera)
	}
}

// DrawUI рисует UI редактора поверх масштабированного изображения
func (le *LevelEditor) DrawUI(scale float32, posX, posY float32) {
	// UI всегда в фиксированной позиции экрана (не зависит от камеры)
	uiPosX := float32(10) // Фиксированная позиция от левого края экрана
	uiPosY := float32(10) // Фиксированная позиция от верхнего края экрана

	// Отрисовка UI
	if le.ShowDropdown {
		dropdownRect := rl.NewRectangle(uiPosX, uiPosY+60, 120, 30)
		raygui.DropdownBox(dropdownRect, "Block;Enemy", &le.DropdownActive, le.ShowDropdown)
	}

	// Текст
	fontSize := int32(20)
	textY := uiPosY + 90
	rl.DrawText("Level Editor: Press Tab to select element, LMB to place, RMB to remove",
		int32(uiPosX), int32(textY), fontSize, rl.Black)

	// Дополнительная информация
	rl.DrawText(fmt.Sprintf("Active: %s", le.getActiveElementName()), int32(uiPosX), int32(textY+30), fontSize, rl.Black)
	rl.DrawText(fmt.Sprintf("Blocks: %d, Enemies: %d", len(le.Level.Blocks), len(le.Level.Enemies)), int32(uiPosX), int32(textY+60), fontSize, rl.Black)
}

// getActiveElementName возвращает название активного элемента
func (le *LevelEditor) getActiveElementName() string {
	switch le.ActiveElementType {
	case BlockElement:
		return "Block"
	case EnemyElement:
		return "Enemy"
	default:
		return "Unknown"
	}
}

// Вспомогательная функция для объединения строк для raygui.DropdownBox
func stringsJoin(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		result += s
		if i < len(strs)-1 {
			result += sep
		}
	}
	return result
}
