package resistor

import (
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	resources = make(map[string]rl.Texture2D)
	models    = make(map[string]rl.Model)
)

// blocksDir is where every placeable level texture lives. assets.go embeds the
// whole directory, so a texture dropped in here ships with the next build.
const blocksDir = "resources/blocks"

// animTable lists every actor animation. The index is the animation's network id,
// so the order must be identical on client and server (append only).
var animTable = []string{
	"resources/threefour/run.png",
	"resources/threefour/idle.png",
	"resources/threefour/jump-start.png",
	"resources/threefour/jump-climb.png",
	"resources/threefour/updown.png",
	"resources/threefour/updown-0-1.png",
	"resources/threefour/updown-0-2.png",
	"resources/threefour/updown-1-0.png",
	"resources/threefour/updown-1-2.png",
	"resources/threefour/updown-2-0.png",
	"resources/threefour/updown-2-1.png",
	"resources/threefour/updown-atk-2-1.png",
	"resources/threefour/updown-atk-2-0.png",
	"resources/threefour/updown-atk-1-2.png",
	"resources/threefour/updown-atk-1-0.png",
	"resources/threefour/updown-atk-0-1.png",
	"resources/threefour/updown-atk-0-2.png",
	"resources/threefour/attack-high.png",
	"resources/threefour/attack-mid.png",
	"resources/threefour/attack-mid-2.png",
	"resources/threefour/attack-low-2.png",
	"resources/threefour/attack-high-2.png",
	"resources/threefour/attack-high-3.png",
	"resources/threefour/attack-low.png",
	"resources/threefour/hit-top.png",
	"resources/threefour/death_fall.png",
	"resources/threefour/back-attack-low-1.png",
	"resources/threefour/back-attack-low-2.png",
	"resources/threefour/back-attack-mid-1.png",
	"resources/threefour/back-attack-high-1.png",
	"resources/threefour/back-attack-high-2.png",
	"resources/threefour/block-all-start.png",
	"resources/threefour/block-all-1.png",
	"resources/threefour/block-low-start.png",
	"resources/threefour/block-low-1.png",
	"resources/threefour/block-mid-start.png",
	"resources/threefour/block-mid-1.png",
	"resources/threefour/block-mid-2.png",
	"resources/threefour/block-mid-v2-start.png",
	"resources/threefour/block-mid-v2-1.png",
	"resources/threefour/block.png",
	"resources/threefour/walk.png",
	"resources/threefour/walk-bot.png",
	"resources/threefour/walk-mid.png",
	"resources/threefour/walk-top.png",
	"resources/threefour/turnover.png",
	"resources/threefour/run-jump.png",
	"resources/threefour/fall.png",
	"resources/threefour/push.png",
	"resources/threefour/battle-low-1-left.png",
	"resources/threefour/battle-low-1-right.png",
}

var animIDs = func() map[string]uint8 {
	m := make(map[string]uint8, len(animTable))
	for i, p := range animTable {
		m[p] = uint8(i)
	}
	return m
}()

func animID(path string) uint8 {
	id, ok := animIDs[path]
	if !ok {
		panic("animation not in animTable: " + path)
	}
	return id
}

func animPath(id uint8) string {
	if int(id) < len(animTable) {
		return animTable[id]
	}
	return animTable[animID("resources/threefour/updown.png")]
}

// embeddedBlocks is every block texture this binary carries. It is what tells
// the editor whether a texture would still be there in a build made now: one
// that is only on disk can be placed, but would not ship.
var embeddedBlocks = func() map[string]bool {
	m := make(map[string]bool)
	for _, p := range embeddedBlockPaths() {
		m[p] = true
	}
	return m
}()

// embeddedBlockPaths lists the block textures in the binary.
func embeddedBlockPaths() []string {
	names, err := fs.Glob(assetFS, blocksDir+"/*.png")
	if err != nil {
		return nil
	}
	return names
}

// diskBlockPaths lists the block textures beside the game on disk. A release
// build never looks: it plays with the textures it was built with.
func diskBlockPaths() []string {
	if !DevTools {
		return nil
	}
	entries, err := os.ReadDir(blocksDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		// the same rule go:embed applies to *.png, so what is listed here is
		// exactly what a rebuild would carry
		n := e.Name()
		if e.IsDir() || path.Ext(n) != ".png" || strings.HasPrefix(n, ".") || strings.HasPrefix(n, "_") {
			continue
		}
		names = append(names, blocksDir+"/"+e.Name())
	}
	return names
}

// mergeTextureLists merges two lists of texture paths into one sorted list with
// no repeats. Sorting is what gives the object picker the same order every run,
// whichever order the two sources came in.
func mergeTextureLists(embedded, disk []string) []string {
	all := make([]string, 0, len(embedded)+len(disk))
	all = append(all, embedded...)
	all = append(all, disk...)
	slices.Sort(all)
	return slices.Compact(all)
}

// blockTextures is every placeable block texture: what the build carries, plus
// whatever a development build finds beside it on disk. It reads no textures, so
// it is safe before there is a window to read them into.
func blockTextures() []string {
	return mergeTextureLists(embeddedBlockPaths(), diskBlockPaths())
}

// inBuild reports whether a texture is embedded in this binary.
func inBuild(name string) bool { return embeddedBlocks[name] }

func LoadTextures() {
	for _, p := range blockTextures() {
		loadBlockTex(p)
	}
	for _, p := range animTable {
		loadTex(p)
	}

	loadTex("resources/main_menu/resistor.png")
	loadModel("resources/main_menu/head.obj")
	loadModel("resources/main_menu/jab1.obj")
	loadModel("resources/main_menu/jab2.obj")
	loadModel("resources/main_menu/kick.obj")
	loadModel("resources/main_menu/martelo.obj")
}

func LoadLevelObjects() []LevelObject {
	return []LevelObject{
		{
			Resource: "resources/blocks/e1m1_wall1.png",
			Layer:    0,
			X:        0,
			Y:        0,
			CollisionRects: []rl.Rectangle{{
				X:      15,
				Y:      86,
				Width:  40,
				Height: 20,
			}},
			LightRects: []rl.Rectangle{{
				X:      22,
				Y:      90,
				Width:  40,
				Height: 15,
			}},
		}, {
			Resource: "resources/blocks/e1m1_door1.png",
			Layer:    0,
			X:        0,
			Y:        0,
			CollisionRects: []rl.Rectangle{{
				X:      0,
				Y:      86,
				Width:  40,
				Height: 20,
			}},
		}, {
			Resource: "resources/blocks/e1m1_platform1.png",
			Layer:    0,
			X:        0,
			Y:        0,
			CollisionRects: []rl.Rectangle{{
				X:      0,
				Y:      86,
				Width:  40,
				Height: 20,
			}},
		},
	}
}

// loadTex decodes an embedded image straight into a GPU texture, so nothing is
// read from disk at runtime.
func loadTex(name string) {
	loadTexBytes(name, asset(name))
}

// loadTexBytes decodes an image into a GPU texture registered under name.
func loadTexBytes(name string, data []byte) bool {
	if !imageDecodable(path.Ext(name), data) {
		rl.TraceLog(rl.LogError, "texture: cannot decode %s", name)
		return false
	}
	img := rl.LoadImageFromMemory(path.Ext(name), data, int32(len(data)))
	resources[name] = rl.LoadTextureFromImage(img)
	texMasks[name] = maskOf(img)
	rl.UnloadImage(img)
	return true
}

// imageDecodable reports whether these bytes are an image raylib can read. It
// decodes on the CPU only, so it can be asked before anything is copied
// anywhere or uploaded to the GPU.
func imageDecodable(ext string, data []byte) bool {
	if len(data) == 0 {
		return false
	}
	img := rl.LoadImageFromMemory(ext, data, int32(len(data)))
	if img == nil || img.Width == 0 || img.Height == 0 {
		return false
	}
	rl.UnloadImage(img)
	return true
}

// loadTexDisk registers a texture read from disk under name. Only a development
// build calls it: a release build has nothing but the assets it was built with.
func loadTexDisk(name, diskPath string) bool {
	if !DevTools {
		return false
	}
	data, err := os.ReadFile(diskPath)
	if err != nil {
		rl.TraceLog(rl.LogError, "texture: read %s: %s", diskPath, err.Error())
		return false
	}
	return loadTexBytes(name, data)
}

// loadBlockTex loads a block texture from wherever it is: the binary, or - for
// one a development build found beside the game - the disk.
func loadBlockTex(name string) {
	if inBuild(name) {
		loadTex(name)
		return
	}
	loadTexDisk(name, name)
}

func loadModel(name string) {
	models[name] = loadOBJ(name, asset(name))
}
