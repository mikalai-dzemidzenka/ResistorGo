package resistor

import (
	"encoding/json"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const settingsPath = "settings.json"

type WindowSize struct {
	W, H int32
}

// windowSizes are the selectable window sizes, all 16:9 so they letterbox to nothing.
var windowSizes = []WindowSize{{1280, 720}, {1600, 900}, {1920, 1080}}

func windowSizesLabel() string { return "1280x720;1600x900;1920x1080" }

type Settings struct {
	Fullscreen  bool `json:"fullscreen"`
	WindowSize  int  `json:"windowSize"` // index into windowSizes
	MusicVolume int  `json:"musicVolume"`
	SoundVolume int  `json:"soundVolume"`
	// HideHUD turns off the in-scene HUD over the fighters - the structure bar,
	// its stun pips and the energy bar - for an unobstructed view of the fight.
	HideHUD bool `json:"hideHud"`
}

func DefaultSettings() Settings {
	return Settings{Fullscreen: false, WindowSize: 0, MusicVolume: 70, SoundVolume: 80, HideHUD: false}
}

var Cfg = DefaultSettings()

// clamp brings every field into range, replacing anything unusable with the default.
func (s *Settings) clamp() {
	d := DefaultSettings()
	if s.WindowSize < 0 || s.WindowSize >= len(windowSizes) {
		s.WindowSize = d.WindowSize
	}
	s.MusicVolume = clamp(s.MusicVolume, 0, 100)
	s.SoundVolume = clamp(s.SoundVolume, 0, 100)
}

// LoadSettings reads settings.json. Anything missing or malformed falls back to
// the defaults: bad settings must never stop the game from starting.
func LoadSettings() {
	Cfg = DefaultSettings()
	b, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}
	s := DefaultSettings()
	if err := json.Unmarshal(b, &s); err != nil {
		return
	}
	s.clamp()
	Cfg = s
}

func SaveSettings() {
	b, err := json.MarshalIndent(Cfg, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(settingsPath, b, 0o644)
}

// ApplySettings makes the current settings take effect: window mode and size,
// music volume, and the letterbox that depends on them.
func ApplySettings() {
	if !rl.IsWindowReady() {
		return
	}
	if Cfg.Fullscreen != rl.IsWindowState(rl.FlagBorderlessWindowedMode) {
		rl.ToggleBorderlessWindowed()
	}
	if !Cfg.Fullscreen {
		ws := windowSizes[Cfg.WindowSize]
		if rl.GetScreenWidth() != int(ws.W) || rl.GetScreenHeight() != int(ws.H) {
			rl.SetWindowSize(int(ws.W), int(ws.H))
			m := rl.GetCurrentMonitor()
			rl.SetWindowPosition(
				(rl.GetMonitorWidth(m)-int(ws.W))/2,
				(rl.GetMonitorHeight(m)-int(ws.H))/2,
			)
		}
	}
	applyMusicVolume()
	letterbox()
}
