package resistor

import (
	"math/rand"
	"path"
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	menuMusicPath  = "resources/music/Neon Sprint.mp3"
	levelMusicPath = "resources/music/Final Run (1).mp3"
	musicFadeSec   = 0.6
	// LevelMusicDuck is how loud the level track plays while the match is held -
	// time manipulation or edit mode - against its full volume.
	LevelMusicDuck = 0.3
)

// punchPaths are the sounds a landed hit makes, one picked at random each time.
var punchPaths = []string{"resources/sounds/punch1.mp3", "resources/sounds/punch2.mp3"}

// blockPaths are the sounds a blow stopped by a guard makes.
var blockPaths = []string{"resources/sounds/block1.mp3"}

// shotPaths are the sounds a pistol shot makes, one picked at random each time.
var shotPaths = []string{"resources/sounds/pistol1.mp3", "resources/sounds/pistol2.mp3"}

// SoundVoices is how many of each sound can play at once. Playing a sound that
// is still playing restarts it, so each sound gets this many aliases sharing
// its samples, and a flurry of blows is heard blow by blow instead of cut short.
const SoundVoices = 3

type voicedSound struct {
	base   rl.Sound
	voices []rl.Sound
	next   int
}

var (
	punches, blocks, shots []*voicedSound
	// its own random source, so hearing a blow can never change what the bots,
	// which draw from the global one, decide
	soundRng = rand.New(rand.NewSource(2))
	// onPunch, onBlock and onShot, when set, are told of every punch, block
	// and shot heard; tests count with them
	onPunch, onBlock, onShot func()
)

// loadSounds decodes the sounds at paths and gives each its voices.
func loadSounds(paths []string) []*voicedSound {
	var set []*voicedSound
	for _, name := range paths {
		data := asset(name)
		wave := rl.LoadWaveFromMemory(path.Ext(name), data, int32(len(data)))
		base := rl.LoadSoundFromWave(wave)
		rl.UnloadWave(wave)
		if !rl.IsSoundValid(base) {
			continue
		}
		s := &voicedSound{base: base, voices: []rl.Sound{base}}
		for range SoundVoices - 1 {
			s.voices = append(s.voices, rl.LoadSoundAlias(base))
		}
		set = append(set, s)
	}
	return set
}

func unloadSounds(set []*voicedSound) {
	for _, s := range set {
		for _, v := range s.voices[1:] {
			rl.UnloadSoundAlias(v)
		}
		rl.UnloadSound(s.base)
	}
}

// nextVoice is the voice of s the next play goes to, round robin.
func (s *voicedSound) nextVoice() int {
	i := s.next
	s.next = (s.next + 1) % len(s.voices)
	return i
}

// playOne plays one sound of set, picked at random, at the sound volume. With
// no audio device - the dedicated server, the tests - set is empty and there is
// nothing to play.
func playOne(set []*voicedSound) {
	if len(set) == 0 {
		return
	}
	s := set[soundRng.Intn(len(set))]
	v := s.voices[s.nextVoice()]
	rl.SetSoundVolume(v, float32(Cfg.SoundVolume)/100)
	rl.PlaySound(v)
}

// punchHeard plays a punch for a landed hit.
func punchHeard() {
	if onPunch != nil {
		onPunch()
	}
	playOne(punches)
}

// blockHeard plays a block for a blow stopped by a guard or by block-all.
func blockHeard() {
	if onBlock != nil {
		onBlock()
	}
	playOne(blocks)
}

// shotHeard plays a gunshot.
func shotHeard() {
	if onShot != nil {
		onShot()
	}
	playOne(shots)
}

// A musicTrack is one looping track, faded in and out by the stage.
type musicTrack struct {
	path  string
	music rl.Music
	ready bool
	fade  float32 // 0..1, times the volume setting
	// rewind makes the track stop at silence rather than pause, so the next
	// time it is wanted it plays from the top.
	rewind  bool
	stopped bool

	// The mp3 decoder streams straight out of the embedded bytes instead of
	// copying them, so the buffer has to stay alive and unmoved until the
	// stream is unloaded.
	data []byte
	pin  runtime.Pinner
}

var (
	menuTrack  = &musicTrack{path: menuMusicPath}
	levelTrack = &musicTrack{path: levelMusicPath, rewind: true}
)

// load opens the track's stream. A track that cannot be loaded stays silent.
func (t *musicTrack) load() {
	t.data = asset(t.path)
	t.pin.Pin(&t.data[0])
	t.music = rl.LoadMusicStreamFromMemory(path.Ext(t.path), t.data, int32(len(t.data)))
	if !rl.IsMusicValid(t.music) {
		t.pin.Unpin()
		t.data = nil
		return
	}
	t.music.Looping = true
	t.ready = true
	t.stopped = true
}

func (t *musicTrack) unload() {
	if !t.ready {
		return
	}
	rl.UnloadMusicStream(t.music)
	t.ready = false
	t.pin.Unpin()
	t.data = nil
}

// initAudio opens the audio device and loads the tracks, starting the menu's.
// Everything is guarded: a machine with no audio output runs silently instead
// of failing.
func initAudio() {
	rl.InitAudioDevice()
	if !rl.IsAudioDeviceReady() {
		return
	}
	menuTrack.load()
	levelTrack.load()
	punches = loadSounds(punchPaths)
	blocks = loadSounds(blockPaths)
	shots = loadSounds(shotPaths)
	menuTrack.fade = 1
	applyMusicVolume()
	if menuTrack.ready {
		rl.PlayMusicStream(menuTrack.music)
		menuTrack.stopped = false
	}
}

func closeAudio() {
	menuTrack.unload()
	levelTrack.unload()
	unloadSounds(punches)
	unloadSounds(blocks)
	unloadSounds(shots)
	punches, blocks, shots = nil, nil, nil
	if rl.IsAudioDeviceReady() {
		rl.CloseAudioDevice()
	}
}

func applyMusicVolume() {
	for _, t := range []*musicTrack{menuTrack, levelTrack} {
		if t.ready {
			rl.SetMusicVolume(t.music, t.fade*float32(Cfg.MusicVolume)/100)
		}
	}
}

// musicTargets is how loud each track is heading, before the volume setting:
// the menu's on the menu, the level's in a match, ducked while the match is
// held.
func musicTargets(stage int, held bool) (menu, level float32) {
	switch stage {
	case StageMainMenu:
		return 1, 0
	case StageLevel:
		if held {
			return 0, LevelMusicDuck
		}
		return 0, 1
	}
	return 0, 0
}

// A playback is what to do to a track's stream this frame.
type playback int

const (
	keepPlaying playback = iota
	keepSilent
	pausePlayback
	stopPlayback
	startPlayback
	resumePlayback
)

// nextPlayback decides what a track's stream does at this fade. A silent track
// pauses, or stops if it rewinds; a track fading in plays from the top if it
// was stopped and resumes if it was paused.
func nextPlayback(rewind, stopped, playing bool, fade float32) playback {
	if fade == 0 {
		switch {
		case rewind && !stopped:
			return stopPlayback
		case !rewind && playing:
			return pausePlayback
		}
		return keepSilent
	}
	switch {
	case stopped:
		return startPlayback
	case !playing:
		return resumePlayback
	}
	return keepPlaying
}

// step fades the track toward target and keeps its stream fed.
func (t *musicTrack) step(dt, target float32) {
	if !t.ready {
		return
	}
	t.fade = moveTo(t.fade, dt/musicFadeSec, target)
	switch nextPlayback(t.rewind, t.stopped, rl.IsMusicStreamPlaying(t.music), t.fade) {
	case stopPlayback:
		rl.StopMusicStream(t.music)
		t.stopped = true
		return
	case pausePlayback:
		rl.PauseMusicStream(t.music)
		return
	case keepSilent:
		return
	case startPlayback:
		rl.PlayMusicStream(t.music)
		t.stopped = false
	case resumePlayback:
		rl.ResumeMusicStream(t.music)
	}
	rl.UpdateMusicStream(t.music)
}

// updateAudio keeps the streams fed and fades the tracks with the stage. It
// must run every frame, in every stage, or the music buffers starve.
func updateAudio(dt float32) {
	menu, level := musicTargets(Stage, matchHeld())
	menuTrack.step(dt, menu)
	levelTrack.step(dt, level)
	applyMusicVolume()
}
