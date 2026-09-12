package resistor

import (
	"path"
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	menuMusicPath = "resources/music/Neon Sprint.mp3"
	musicFadeSec  = 0.6
)

var (
	menuMusic  rl.Music
	musicReady bool
	musicFade  float32 // 1 in the menu, 0 in a match

	// The mp3 decoder streams straight out of the embedded bytes instead of
	// copying them, so the buffer has to stay alive and unmoved until the
	// stream is unloaded.
	menuMusicData []byte
	menuMusicPin  runtime.Pinner
)

// initAudio opens the audio device and loads the menu track. Everything is
// guarded: a machine with no audio output runs silently instead of failing.
func initAudio() {
	rl.InitAudioDevice()
	if !rl.IsAudioDeviceReady() {
		return
	}
	menuMusicData = asset(menuMusicPath)
	menuMusicPin.Pin(&menuMusicData[0])
	menuMusic = rl.LoadMusicStreamFromMemory(path.Ext(menuMusicPath), menuMusicData, int32(len(menuMusicData)))
	if !rl.IsMusicValid(menuMusic) {
		menuMusicPin.Unpin()
		menuMusicData = nil
		return
	}
	menuMusic.Looping = true
	musicReady = true
	applyMusicVolume()
	rl.PlayMusicStream(menuMusic)
}

func closeAudio() {
	if musicReady {
		rl.UnloadMusicStream(menuMusic)
		musicReady = false
		menuMusicPin.Unpin()
		menuMusicData = nil
	}
	if rl.IsAudioDeviceReady() {
		rl.CloseAudioDevice()
	}
}

func applyMusicVolume() {
	if !musicReady {
		return
	}
	rl.SetMusicVolume(menuMusic, musicFade*float32(Cfg.MusicVolume)/100)
}

// updateAudio keeps the stream fed and fades the music with the stage. It must
// run every frame, in every stage, or the music buffer starves.
func updateAudio(dt float32) {
	if !musicReady {
		return
	}
	want := float32(0)
	if Stage == StageMainMenu {
		want = 1
	}
	musicFade = moveTo(musicFade, dt/musicFadeSec, want)
	applyMusicVolume()

	if musicFade == 0 {
		if rl.IsMusicStreamPlaying(menuMusic) {
			rl.PauseMusicStream(menuMusic)
		}
		return
	}
	if !rl.IsMusicStreamPlaying(menuMusic) {
		rl.ResumeMusicStream(menuMusic)
	}
	rl.UpdateMusicStream(menuMusic)
}
