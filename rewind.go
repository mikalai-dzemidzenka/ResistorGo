package resistor

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	// RewindFrames is how many fixed steps of history every actor keeps.
	RewindFrames = 256
	// ScrubFrames is how many recorded frames one fixed step of scrubbing
	// covers. It is the same in both directions, so 2 means time runs at twice
	// the speed it was recorded at whichever way it is going.
	ScrubFrames = 2
	// ScrubDeadZone is the horizontal mouse movement, in pixels for one frame,
	// below which the mouse counts as still. Anything the hand actually does
	// moves at least a whole pixel; this only swallows sub-pixel sensor noise.
	ScrubDeadZone = 0.5
)

// In a single-player match TimeKey toggles time manipulation and RestartKey
// starts the level over.
const (
	TimeKey    = rl.KeyT
	RestartKey = rl.KeyR
)

// Time-manipulation mode. Off is normal play; on, the match is either held on
// one recorded frame or travelling through the recording.
const (
	TimeOff = iota
	TimePaused
	TimeScrubbing
)

var (
	TimeMode = TimeOff
	// ScrubDir is -1 while rewinding, +1 while replaying forward, 0 while the
	// mouse is still.
	ScrubDir int
	scrubAcc float64
)

// The recorded history is a window on a ring buffer: rewindHead is the newest
// frame recorded, rewindCount how many frames behind it are still valid, and
// rewindCursor the frame currently on screen. Frame numbers are absolute and
// only ever grow; rewindSlot maps them onto the buffer.
var (
	rewindHead   int
	rewindCount  int
	rewindCursor int
)

// A worldFrame is the whole single-player match as it stood after one fixed
// step. It keeps the fighters themselves as well as a copy of each one's value:
// a fighter object is never reused once it leaves the world, so every pointer
// in a recorded value names an object this frame also records, or one already
// gone that the live match held just the same. Putting every value back is
// then the whole restore, with nothing to relink.
type worldFrame struct {
	tick    uint32
	nextID  uint16
	local   *Actor
	pending []pendingSpawn
	dying   []*Actor
	actors  []actorFrame
}

type actorFrame struct {
	a *Actor
	v Actor
}

// rewindFrames is the recording itself, indexed by rewindSlot. Each slot's
// slices are reused from one lap of the ring to the next.
var rewindFrames [RewindFrames]worldFrame

// captureFrame records w into the frame the recording is open on.
func captureFrame(w *World) {
	f := &rewindFrames[rewindSlot(rewindHead)]
	f.tick, f.nextID, f.local = w.Tick, w.nextID, w.Local
	f.pending = append(f.pending[:0], w.pending...)
	f.dying = append(f.dying[:0], w.dying...)
	clear(f.dying[len(f.dying):cap(f.dying)])
	f.actors = f.actors[:0]
	for _, a := range w.Actors {
		f.actors = append(f.actors, actorFrame{a, *a})
	}
	clear(f.actors[len(f.actors):cap(f.actors)]) // a lap with fewer fighters holds none of the last lap's
}

// restoreFrame puts w back onto the frame the playhead is on: every fighter
// that was in the world then, as it was then, and no other. A recorded frame
// is not in motion, and the sounds and flashes of what happened on it have
// already played.
func restoreFrame(w *World) {
	f := &rewindFrames[rewindSlot(rewindCursor)]
	w.Tick, w.nextID, w.Local = f.tick, f.nextID, f.local
	w.pending = append(w.pending[:0], f.pending...)
	w.dying = append(w.dying[:0], f.dying...)
	w.Actors = w.Actors[:0]
	for _, r := range f.actors {
		*r.a = r.v
		r.a.PrevPos, r.a.InterpolatedPos = r.a.Pos, r.a.Pos
		r.a.Events = 0
		w.Actors = append(w.Actors, r.a)
	}
}

// rewindSlot maps an absolute frame number onto the ring buffer. Go keeps the
// sign of the left operand in %, so a frame number below zero needs a wrap of
// its own or it indexes the buffer backwards.
func rewindSlot(frame int) int {
	slot := frame % RewindFrames
	if slot < 0 {
		slot += RewindFrames
	}
	return slot
}

// rewindOldest is the oldest frame still held.
func rewindOldest() int {
	if rewindCount == 0 {
		return rewindHead
	}
	return rewindHead - rewindCount + 1
}

// rewindRecord opens the next frame for recording and puts the playhead on it.
// Actors write their state into the slot of the new head.
func rewindRecord() {
	if rewindCount > 0 {
		rewindHead++
	}
	if rewindCount < RewindFrames {
		rewindCount++
	}
	rewindCursor = rewindHead
}

// rewindSeek moves the playhead by delta frames, clamped to the recorded
// history, and reports whether it actually moved.
func rewindSeek(delta int) bool {
	was := rewindCursor
	rewindCursor = clamp(rewindCursor+delta, rewindOldest(), rewindHead)
	return rewindCursor != was
}

// rewindResumeHere makes the frame on screen the newest one: the frames ahead
// of it record a future that is not going to happen now.
func rewindResumeHere() {
	rewindCount -= rewindHead - rewindCursor
	rewindHead = rewindCursor
}

// TimeShowsHistory reports whether what is on screen is a recorded frame
// rather than a live one, which is all the actors need to know.
func TimeShowsHistory() bool {
	return TimeMode != TimeOff && rewindCount > 0
}

// ResetTime leaves time-manipulation mode and forgets the recording. A new
// match starts with no history, and holds none of the old match's fighters.
func ResetTime() {
	TimeMode, ScrubDir, scrubAcc = TimeOff, 0, 0
	rewindHead, rewindCount, rewindCursor = 0, 0, 0
	rewindFrames = [RewindFrames]worldFrame{}
}

// scrubDirection turns one frame of horizontal mouse movement into a direction
// of travel. Only the direction matters: the rate is fixed, so how fast the
// hand moves does not change how fast time does.
func scrubDirection(dx float32) int {
	switch {
	case dx > ScrubDeadZone:
		return 1
	case dx < -ScrubDeadZone:
		return -1
	}
	return 0
}

// UpdateTime advances time manipulation by one rendered frame: TimeKey toggles the
// mode, and while it is on the horizontal mouse movement dx scrubs through the
// recording. It reports whether the playhead moved, in which case the caller
// restores the world onto it. Input is passed in rather than read here so the
// state machine can be tested without a window.
func UpdateTime(pressed bool, dx, dt float32) bool {
	if Editing {
		pressed = false // edit mode holds the match; time manipulation waits
	}
	if pressed {
		if TimeMode == TimeOff {
			TimeMode = TimePaused
		} else {
			// Resuming here: the frames ahead are a future that will not happen,
			// and a click made while the match was held is not an attack now.
			rewindResumeHere()
			TimeMode = TimeOff
			clearLatchedInput()
		}
		ScrubDir, scrubAcc = 0, 0
		return false
	}
	if TimeMode == TimeOff {
		return false
	}
	if ScrubDir = scrubDirection(dx); ScrubDir == 0 {
		TimeMode = TimePaused
		scrubAcc = 0
		return false
	}
	TimeMode = TimeScrubbing

	// Scrub in the same fixed quanta the simulation runs in, so the rate does
	// not depend on the frame rate.
	scrubAcc += float64(dt)
	moved := false
	for scrubAcc >= FixedTimeStep {
		scrubAcc -= FixedTimeStep
		moved = rewindSeek(ScrubDir*ScrubFrames) || moved
	}
	return moved
}

// timeModeName labels the mode for the debug overlay.
func timeModeName() string {
	switch {
	case TimeMode == TimeOff:
		return "live"
	case TimeMode == TimeScrubbing && ScrubDir < 0:
		return "rewinding"
	case TimeMode == TimeScrubbing:
		return "forward"
	}
	return "paused"
}
