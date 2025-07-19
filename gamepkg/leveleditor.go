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

type BlockType int

const (
	NormalBlock BlockType = iota
	GrassBlock
	StoneBlock
	WoodBlock
	MetalBlock
	BackgroundBlock // Для параллакса
)

type Block struct {
	Position rl.Vector2
	Size     rl.Vector2
	Type     BlockType
}

// EnemyData упрощенная структура для сохранения врагов в JSON
type EnemyData struct {
	Position rl.Vector2
	Type     ActorType
}

type Level struct {
	Name    string
	Blocks  []Block
	Enemies []EnemyData
}

type LevelManager struct {
	Levels       []Level
	CurrentLevel int
}

type LevelEditor struct {
	ActiveElementType LevelElementType
	ActiveBlockType   BlockType
	ShowDropdown      bool
	DropdownActive    int32
	LevelManager      LevelManager
	DisplayEnemies    []Actor         // Враги для отображения в редакторе
	TextureManager    *TextureManager // Менеджер текстур
	// Новое поле для управления раскрытием DropdownBox
	dropdownEditMode bool
}

func NewLevelManager() LevelManager {
	return LevelManager{
		Levels:       []Level{},
		CurrentLevel: 0,
	}
}

func (lm *LevelManager) AddLevel(level Level) {
	lm.Levels = append(lm.Levels, level)
}

func (lm *LevelManager) GetCurrentLevel() *Level {
	if len(lm.Levels) == 0 {
		return nil
	}
	return &lm.Levels[lm.CurrentLevel]
}

func (lm *LevelManager) NextLevel() {
	if len(lm.Levels) > 0 {
		lm.CurrentLevel = (lm.CurrentLevel + 1) % len(lm.Levels)
	}
}

func (lm *LevelManager) PreviousLevel() {
	if len(lm.Levels) > 0 {
		lm.CurrentLevel = (lm.CurrentLevel - 1 + len(lm.Levels)) % len(lm.Levels)
	}
}

func (lm *LevelManager) SaveAllLevels() error {
	for i, level := range lm.Levels {
		filename := fmt.Sprintf("level_%d.json", i)
		data, err := json.MarshalIndent(level, "", "  ")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (lm *LevelManager) LoadAllLevels() error {
	lm.Levels = []Level{}

	// Загружаем все доступные уровни
	for i := 0; i < 10; i++ { // Максимум 10 уровней
		filename := fmt.Sprintf("level_%d.json", i)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			// Если файл не найден, останавливаемся
			break
		}

		var level Level
		err = json.Unmarshal(data, &level)
		if err != nil {
			fmt.Printf("Error loading level %d: %v\n", i, err)
			continue
		}

		lm.Levels = append(lm.Levels, level)
	}

	// Если нет уровней, создаём демо-уровни
	if len(lm.Levels) == 0 {
		lm.GenerateDemoLevels()
	}

	return nil
}

func (lm *LevelManager) GenerateDemoLevels() {
	// Уровень 1: Простой
	level1 := Level{
		Name: "Tutorial",
		Blocks: []Block{
			{Position: rl.NewVector2(100, 350), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(148, 350), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(196, 350), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(500, 350), Size: rl.NewVector2(48, 48), Type: StoneBlock},
			{Position: rl.NewVector2(548, 350), Size: rl.NewVector2(48, 48), Type: StoneBlock},
		},
		Enemies: []EnemyData{
			{Position: rl.NewVector2(600, 360), Type: AIActor},
		},
	}

	// Уровень 2: Средний
	level2 := Level{
		Name: "Platforms",
		Blocks: []Block{
			{Position: rl.NewVector2(50, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(98, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(200, 300), Size: rl.NewVector2(48, 48), Type: WoodBlock},
			{Position: rl.NewVector2(248, 300), Size: rl.NewVector2(48, 48), Type: WoodBlock},
			{Position: rl.NewVector2(350, 250), Size: rl.NewVector2(48, 48), Type: StoneBlock},
			{Position: rl.NewVector2(398, 250), Size: rl.NewVector2(48, 48), Type: StoneBlock},
			{Position: rl.NewVector2(500, 350), Size: rl.NewVector2(48, 48), Type: MetalBlock},
			{Position: rl.NewVector2(548, 350), Size: rl.NewVector2(48, 48), Type: MetalBlock},
		},
		Enemies: []EnemyData{
			{Position: rl.NewVector2(250, 260), Type: AIActor},
			{Position: rl.NewVector2(550, 360), Type: AIActor},
		},
	}

	// Уровень 3: Сложный
	level3 := Level{
		Name: "Arena",
		Blocks: []Block{
			{Position: rl.NewVector2(0, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(48, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(96, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(144, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(192, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(240, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(288, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(336, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(384, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(432, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(480, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(528, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(576, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(624, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(672, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(720, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},
			{Position: rl.NewVector2(768, 400), Size: rl.NewVector2(48, 48), Type: GrassBlock},

			// Платформы
			{Position: rl.NewVector2(150, 300), Size: rl.NewVector2(48, 48), Type: WoodBlock},
			{Position: rl.NewVector2(198, 300), Size: rl.NewVector2(48, 48), Type: WoodBlock},
			{Position: rl.NewVector2(400, 250), Size: rl.NewVector2(48, 48), Type: StoneBlock},
			{Position: rl.NewVector2(448, 250), Size: rl.NewVector2(48, 48), Type: StoneBlock},
			{Position: rl.NewVector2(600, 300), Size: rl.NewVector2(48, 48), Type: MetalBlock},
			{Position: rl.NewVector2(648, 300), Size: rl.NewVector2(48, 48), Type: MetalBlock},
		},
		Enemies: []EnemyData{
			{Position: rl.NewVector2(200, 260), Type: AIActor},
			{Position: rl.NewVector2(450, 210), Type: AIActor},
			{Position: rl.NewVector2(650, 260), Type: AIActor},
		},
	}

	lm.Levels = []Level{level1, level2, level3}
	lm.CurrentLevel = 0
}

func NewLevelEditor() LevelEditor {
	levelManager := NewLevelManager()
	levelManager.LoadAllLevels()
	textureManager := NewTextureManager()

	return LevelEditor{
		ActiveElementType: BlockElement,
		ActiveBlockType:   NormalBlock,
		ShowDropdown:      false,
		DropdownActive:    0,
		LevelManager:      levelManager,
		TextureManager:    textureManager,
	}
}

func (le *LevelEditor) SaveCurrentLevel() error {
	currentLevel := le.LevelManager.GetCurrentLevel()
	if currentLevel == nil {
		return fmt.Errorf("no current level")
	}

	filename := fmt.Sprintf("level_%d.json", le.LevelManager.CurrentLevel)
	data, err := json.MarshalIndent(currentLevel, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (le *LevelEditor) LoadLevel(filename string) error {
	// Устаревший метод, оставляем для совместимости
	return nil
}

func (le *LevelEditor) Update(camera *rl.Camera2D, scale float32, posX, posY float32) {
	// Переключение типа блока через B
	if rl.IsKeyPressed(rl.KeyB) && le.DropdownActive == 0 {
		le.ActiveBlockType = BlockType((int(le.ActiveBlockType) + 1) % 6)
	}

	items := []string{"Block", "Enemy"}
	uiPosX := float32(10)
	uiPosY := float32(10)
	dropdownRect := rl.NewRectangle(uiPosX, uiPosY+60, 120, 30)

	// Клик по списку — раскрыть/закрыть
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), dropdownRect) {
		le.dropdownEditMode = !le.dropdownEditMode
	}

	// Если раскрыт и выбран пункт — обновить выбор и закрыть
	if raygui.DropdownBox(dropdownRect, stringsJoin(items, ";"), &le.DropdownActive, le.dropdownEditMode) {
		le.dropdownEditMode = false
	}
	le.ActiveElementType = LevelElementType(le.DropdownActive)

	// Клик вне списка — закрыть
	if le.dropdownEditMode && rl.IsMouseButtonPressed(rl.MouseLeftButton) && !rl.CheckCollisionPointRec(rl.GetMousePosition(), dropdownRect) {
		le.dropdownEditMode = false
	}

	// Размещение элемента (ЛКМ)
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !le.dropdownEditMode {
		mousePos := rl.GetMousePosition()
		if mousePos.X >= posX && mousePos.X <= posX+float32(ScreenWidth)*scale &&
			mousePos.Y >= posY && mousePos.Y <= posY+float32(ScreenHeight)*scale {
			normalizedX := (mousePos.X - posX) / (float32(ScreenWidth) * scale)
			normalizedY := (mousePos.Y - posY) / (float32(ScreenHeight) * scale)
			sceneX := normalizedX * float32(ScreenWidth)
			sceneY := normalizedY * float32(ScreenHeight)
			worldPos := rl.GetScreenToWorld2D(rl.NewVector2(sceneX, sceneY), *camera)
			currentLevel := le.LevelManager.GetCurrentLevel()
			if currentLevel == nil {
				return
			}
			saveNeeded := false
			switch le.DropdownActive {
			case 0: // Block
				block := Block{
					Position: rl.NewVector2(float32(int(worldPos.X/48)*48), float32(int(worldPos.Y/48)*48)),
					Size:     rl.NewVector2(48*SpriteScale, 48*SpriteScale),
					Type:     le.ActiveBlockType,
				}
				currentLevel.Blocks = append(currentLevel.Blocks, block)
				saveNeeded = true
			case 1: // Enemy
				enemyData := EnemyData{
					Position: rl.NewVector2(worldPos.X, ScreenHeight-90*SpriteScale),
					Type:     AIActor,
				}
				currentLevel.Enemies = append(currentLevel.Enemies, enemyData)
				saveNeeded = true
			}
			if saveNeeded {
				le.SaveCurrentLevel()
			}
		}
	}

	// Удаление элемента (ПКМ)
	if rl.IsMouseButtonPressed(rl.MouseRightButton) && !le.dropdownEditMode {
		mousePos := rl.GetMousePosition()
		if mousePos.X >= posX && mousePos.X <= posX+float32(ScreenWidth)*scale &&
			mousePos.Y >= posY && mousePos.Y <= posY+float32(ScreenHeight)*scale {
			normalizedX := (mousePos.X - posX) / (float32(ScreenWidth) * scale)
			normalizedY := (mousePos.Y - posY) / (float32(ScreenHeight) * scale)
			sceneX := normalizedX * float32(ScreenWidth)
			sceneY := normalizedY * float32(ScreenHeight)
			worldPos := rl.GetScreenToWorld2D(rl.NewVector2(sceneX, sceneY), *camera)
			currentLevel := le.LevelManager.GetCurrentLevel()
			if currentLevel == nil {
				return
			}
			// Удаление блоков
			newBlocks := []Block{}
			for _, block := range currentLevel.Blocks {
				blockRect := rl.NewRectangle(block.Position.X, block.Position.Y, block.Size.X, block.Size.Y)
				if !rl.CheckCollisionPointRec(worldPos, blockRect) {
					newBlocks = append(newBlocks, block)
				}
			}
			currentLevel.Blocks = newBlocks
			// Удаление врагов
			newEnemies := []EnemyData{}
			for _, enemy := range currentLevel.Enemies {
				enemyRect := rl.NewRectangle(enemy.Position.X, enemy.Position.Y, 48*SpriteScale, 48*SpriteScale)
				if !rl.CheckCollisionPointRec(worldPos, enemyRect) {
					newEnemies = append(newEnemies, enemy)
				}
			}
			currentLevel.Enemies = newEnemies
			le.SaveCurrentLevel()
		}
	}
}

// DrawScene рисует только сцену редактора (без UI)
func (le *LevelEditor) DrawScene(camera *rl.Camera2D) {
	currentLevel := le.LevelManager.GetCurrentLevel()
	if currentLevel == nil {
		return
	}

	// Отрисовка блоков с разными цветами и параллаксом
	for _, block := range currentLevel.Blocks {
		var color rl.Color
		var parallaxOffset float32 = 0

		switch block.Type {
		case NormalBlock:
			color = rl.Gray
		case GrassBlock:
			color = rl.Green
			parallaxOffset = camera.Target.X * 0.1 // Медленный параллакс
		case StoneBlock:
			color = rl.DarkGray
			parallaxOffset = camera.Target.X * 0.05 // Очень медленный параллакс
		case WoodBlock:
			color = rl.Brown
			parallaxOffset = camera.Target.X * 0.15 // Средний параллакс
		case MetalBlock:
			color = rl.LightGray
		case BackgroundBlock:
			color = rl.Fade(rl.Blue, 0.3)          // Полупрозрачный
			parallaxOffset = camera.Target.X * 0.3 // Быстрый параллакс
		}

		// Применяем параллакс
		parallaxPos := rl.NewVector2(block.Position.X+parallaxOffset, block.Position.Y)
		rl.DrawRectangleV(parallaxPos, block.Size, color)
	}

	// Обновляем кэш врагов для отображения
	le.updateDisplayEnemies(currentLevel)

	// Отрисовка врагов из кэша
	for i := range le.DisplayEnemies {
		DrawActor(&le.DisplayEnemies[i], nil, rl.Blue, camera)
	}
}

// updateDisplayEnemies обновляет кэш врагов для отображения
func (le *LevelEditor) updateDisplayEnemies(level *Level) {
	// Если количество врагов изменилось, пересоздаем кэш
	if len(le.DisplayEnemies) != len(level.Enemies) {
		// Освобождаем старые текстуры
		for i := range le.DisplayEnemies {
			UnloadActorTextures(&le.DisplayEnemies[i])
		}

		// Создаем новые акторы
		le.DisplayEnemies = make([]Actor, len(level.Enemies))
		for i, enemyData := range level.Enemies {
			le.DisplayEnemies[i] = NewActor(enemyData.Position, enemyData.Type, le.TextureManager)
		}
	} else {
		// Обновляем позиции существующих акторов
		for i, enemyData := range level.Enemies {
			le.DisplayEnemies[i].Position = enemyData.Position
		}
	}
}

func (le *LevelEditor) DrawUI(scale float32, posX, posY float32) {
	uiPosX := float32(10)
	uiPosY := float32(10)
	fontSize := int32(20)
	textY := uiPosY + 90

	// UI выбора типа элемента всегда видим
	items := []string{"Block", "Enemy"}
	dropdownRect := rl.NewRectangle(uiPosX, uiPosY+60, 120, 30)
	raygui.DropdownBox(dropdownRect, stringsJoin(items, ";"), &le.DropdownActive, false)
	// Показываем текущий активный элемент
	rl.DrawText(fmt.Sprintf("Active: %s", le.getActiveElementName()), int32(uiPosX), int32(textY+90), fontSize, rl.Black)
	if le.ActiveElementType == BlockElement {
		blockTypeName := le.getBlockTypeName()
		rl.DrawText(fmt.Sprintf("Block Type: %s", blockTypeName), int32(uiPosX), int32(textY+120), fontSize, rl.Black)
	}
	// Остальной UI
	rl.DrawText("Level Editor: Tab=select, B=block type, LMB=place, RMB=remove", int32(uiPosX), int32(textY), fontSize, rl.Black)
	rl.DrawText("Arrow Keys: Switch levels", int32(uiPosX), int32(textY+30), fontSize, rl.Black)
	currentLevel := le.LevelManager.GetCurrentLevel()
	if currentLevel != nil {
		rl.DrawText(fmt.Sprintf("Level: %s (%d/%d)", currentLevel.Name, le.LevelManager.CurrentLevel+1, len(le.LevelManager.Levels)), int32(uiPosX), int32(textY+60), fontSize, rl.Black)
		rl.DrawText(fmt.Sprintf("Blocks: %d, Enemies: %d", len(currentLevel.Blocks), len(currentLevel.Enemies)), int32(uiPosX), int32(textY+150), fontSize, rl.Black)
	}
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

// getBlockTypeName возвращает название типа блока
func (le *LevelEditor) getBlockTypeName() string {
	switch le.ActiveBlockType {
	case NormalBlock:
		return "Normal"
	case GrassBlock:
		return "Grass"
	case StoneBlock:
		return "Stone"
	case WoodBlock:
		return "Wood"
	case MetalBlock:
		return "Metal"
	case BackgroundBlock:
		return "Background"
	default:
		return "Unknown"
	}
}

// Cleanup освобождает ресурсы редактора уровней
func (le *LevelEditor) Cleanup() {
	// Освобождаем текстуры всех врагов в кэше
	for i := range le.DisplayEnemies {
		UnloadActorTextures(&le.DisplayEnemies[i])
	}
	le.DisplayEnemies = nil
	if le.TextureManager != nil {
		le.TextureManager.UnloadAll()
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
