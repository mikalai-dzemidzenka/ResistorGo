package resistor

import (
	"os"
	"path/filepath"
	"testing"
)

// inTempDir runs f with the working directory set to a fresh temp dir, so the
// test never touches the real settings.json.
func inTempDir(t *testing.T, f func()) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)
	f()
}

func TestSettingsDefaults(t *testing.T) {
	inTempDir(t, func() {
		LoadSettings() // no file present
		if Cfg != DefaultSettings() {
			t.Errorf("missing file gave %+v, want defaults %+v", Cfg, DefaultSettings())
		}
	})
}

func TestSettingsMalformed(t *testing.T) {
	inTempDir(t, func() {
		os.WriteFile(settingsPath, []byte("{not json at all"), 0o644)
		LoadSettings()
		if Cfg != DefaultSettings() {
			t.Errorf("malformed file gave %+v, want defaults %+v", Cfg, DefaultSettings())
		}
	})
}

func TestSettingsClamped(t *testing.T) {
	cases := []struct {
		name     string
		json     string
		wantSize int
		wantVol  int
	}{
		{"size too high", `{"windowSize":99,"musicVolume":50}`, 0, 50},
		{"size negative", `{"windowSize":-3,"musicVolume":50}`, 0, 50},
		{"volume too high", `{"windowSize":1,"musicVolume":900}`, 1, 100},
		{"volume negative", `{"windowSize":1,"musicVolume":-20}`, 1, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			inTempDir(t, func() {
				os.WriteFile(settingsPath, []byte(c.json), 0o644)
				LoadSettings()
				if Cfg.WindowSize != c.wantSize || Cfg.MusicVolume != c.wantVol {
					t.Errorf("got size %d vol %d, want size %d vol %d",
						Cfg.WindowSize, Cfg.MusicVolume, c.wantSize, c.wantVol)
				}
			})
		})
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	inTempDir(t, func() {
		Cfg = Settings{Fullscreen: true, WindowSize: 2, MusicVolume: 30}
		want := Cfg
		SaveSettings()
		if _, err := os.Stat(filepath.Join(".", settingsPath)); err != nil {
			t.Fatalf("settings file not written: %v", err)
		}
		Cfg = DefaultSettings()
		LoadSettings()
		if Cfg != want {
			t.Errorf("round trip gave %+v, want %+v", Cfg, want)
		}
	})
}

// The sound volume defaults to 80, is kept in range, survives a save, and a
// settings file from before it existed gives the default with the rest as saved.
func TestSoundVolumeSetting(t *testing.T) {
	if d := DefaultSettings().SoundVolume; d != 80 {
		t.Errorf("default sound volume %d, want 80", d)
	}
	for _, c := range []struct {
		name string
		json string
		want Settings
	}{
		{"older file", `{"fullscreen":true,"windowSize":1,"musicVolume":40}`, Settings{Fullscreen: true, WindowSize: 1, MusicVolume: 40, SoundVolume: 80}},
		{"too loud", `{"soundVolume":300}`, func() Settings { s := DefaultSettings(); s.SoundVolume = 100; return s }()},
		{"negative", `{"soundVolume":-5}`, func() Settings { s := DefaultSettings(); s.SoundVolume = 0; return s }()},
	} {
		t.Run(c.name, func(t *testing.T) {
			inTempDir(t, func() {
				os.WriteFile(settingsPath, []byte(c.json), 0o644)
				LoadSettings()
				if Cfg != c.want {
					t.Errorf("got %+v, want %+v", Cfg, c.want)
				}
			})
		})
	}
	inTempDir(t, func() {
		Cfg = DefaultSettings()
		Cfg.SoundVolume = 25
		SaveSettings()
		Cfg = DefaultSettings()
		LoadSettings()
		if Cfg.SoundVolume != 25 {
			t.Errorf("sound volume came back as %d, want 25", Cfg.SoundVolume)
		}
	})
	Cfg = DefaultSettings()
}
