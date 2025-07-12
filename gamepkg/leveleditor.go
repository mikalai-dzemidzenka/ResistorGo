package gamepkg

import (
	"encoding/json"
	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"io/ioutil"
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

type Level struct {
	Blocks  []Block
	Enemies []Player
}

type LevelEditor struct {
	ActiveElementType LevelElementType
	ShowDropdown      bool
	DropdownActive    int32
	Level             Level
}

func NewLevelEditor() LevelEditor {
	return LevelEditor{
		ActiveElementType: BlockElement,
		ShowDropdown:      false,
		DropdownActive:    0,
		Level: Level{
			Blocks:  []Block{},
			Enemies: []Player{},
		},
	}
}

func (le *LevelEditor) SaveLevel(filename string) error {
	data, err := json.MarshalIndent(le.Level, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (le *LevelEditor) LoadLevel(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		// Если файла нет, инициализируем пустой уровень
		le.Level = Level{
			Blocks:  []Block{},
			Enemies: []Player{},
		}
		return nil
	}
	return json.Unmarshal(data, &le.Level)
}

func (le *LevelEditor) Update(camera *rl.Camera2D) {
	// Переключение типа элемента через Tab
	if rl.IsKeyPressed(rl.KeyTab) {
		le.ShowDropdown = !le.ShowDropdown
	}

	// Обработка UI (выпадающий список)
	if le.ShowDropdown {
		items := []string{"Block", "Enemy"}
		if raygui.DropdownBox(rl.NewRectangle(10, 70, 120, 30), stringsJoin(items, ";"), &le.DropdownActive, le.ShowDropdown) {
			le.ActiveElementType = LevelElementType(le.DropdownActive)
			le.ShowDropdown = false
		}
	}

	// Размещение элемента (ЛКМ)
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !le.ShowDropdown {
		mousePos := rl.GetScreenToWorld2D(rl.GetMousePosition(), *camera)
		switch le.ActiveElementType {
		case BlockElement:
			block := Block{
				Position: rl.NewVector2(float32(int(mousePos.X/48)*48), float32(int(mousePos.Y/48)*48)), // Привязка к сетке 48x48
				Size:     rl.NewVector2(48*SpriteScale, 48*SpriteScale),
			}
			le.Level.Blocks = append(le.Level.Blocks, block)
		case EnemyElement:
			enemy := NewPlayer(rl.NewVector2(mousePos.X, ScreenHeight-90*SpriteScale), false)
			le.Level.Enemies = append(le.Level.Enemies, enemy)
		}
	}

	// Удаление элемента (ПКМ)
	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		mousePos := rl.GetScreenToWorld2D(rl.GetMousePosition(), *camera)
		// Удаление блоков
		newBlocks := []Block{}
		for _, block := range le.Level.Blocks {
			blockRect := rl.NewRectangle(block.Position.X, block.Position.Y, block.Size.X, block.Size.Y)
			if !rl.CheckCollisionPointRec(mousePos, blockRect) {
				newBlocks = append(newBlocks, block)
			}
		}
		le.Level.Blocks = newBlocks

		// Удаление врагов
		newEnemies := []Player{}
		for _, enemy := range le.Level.Enemies {
			enemyRect := rl.NewRectangle(enemy.Position.X, enemy.Position.Y, enemy.Size.X, enemy.Size.Y)
			if !rl.CheckCollisionPointRec(mousePos, enemyRect) {
				newEnemies = append(newEnemies, enemy)
			}
		}
		le.Level.Enemies = newEnemies
	}
}

func (le *LevelEditor) Draw(camera *rl.Camera2D) {
	// Отрисовка блоков
	for _, block := range le.Level.Blocks {
		rl.DrawRectangleV(block.Position, block.Size, rl.Gray)
	}

	// Отрисовка врагов
	for i := range le.Level.Enemies {
		DrawPlayer(&le.Level.Enemies[i], nil, rl.Blue, false, camera)
	}

	// Отрисовка UI
	if le.ShowDropdown {
		raygui.DropdownBox(rl.NewRectangle(10, 70, 120, 30), "Block;Enemy", &le.DropdownActive, le.ShowDropdown)
	}
	rl.DrawText("Level Editor: Press Tab to select element, LMB to place, RMB to remove", 10, 100, 20, rl.Black)
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
