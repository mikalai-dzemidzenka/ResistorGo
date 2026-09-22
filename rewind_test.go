package resistor

import "testing"

// resetRewind puts the recording back to an empty match.
func resetRewind() {
	rewindHead, rewindCount, rewindCursor = 0, 0, 0
}

// record n frames of history, as a match would.
func recordFrames(n int) {
	for i := 0; i < n; i++ {
		rewindRecord()
	}
}

// Frame numbers are absolute and the buffer is a ring: every frame number,
// including one below zero, must land inside the buffer.
func TestRewindSlot(t *testing.T) {
	for _, frame := range []int{-RewindFrames - 1, -RewindFrames, -1, 0, 1, RewindFrames - 1, RewindFrames, RewindFrames*3 + 7} {
		if slot := rewindSlot(frame); slot < 0 || slot >= RewindFrames {
			t.Errorf("frame %d maps to slot %d, outside [0,%d)", frame, slot, RewindFrames)
		}
	}
	if a, b := rewindSlot(-1), rewindSlot(RewindFrames-1); a != b {
		t.Errorf("frame -1 maps to %d and frame %d to %d, want the same slot", a, RewindFrames-1, b)
	}
}

// Recording advances the window and leaves the playhead on the newest frame.
func TestRewindRecord(t *testing.T) {
	resetRewind()
	recordFrames(1)
	if rewindCount != 1 || rewindCursor != rewindHead {
		t.Errorf("after one frame: count %d cursor %d head %d, want count 1 and the cursor on the head", rewindCount, rewindCursor, rewindHead)
	}
	recordFrames(99)
	if rewindCount != 100 {
		t.Errorf("after 100 frames: count %d, want 100", rewindCount)
	}
	if rewindCursor != rewindHead {
		t.Errorf("the playhead is at %d, want it on the head %d", rewindCursor, rewindHead)
	}

	recordFrames(RewindFrames * 2) // far past the end of the buffer
	if rewindCount != RewindFrames {
		t.Errorf("count grew to %d, want it capped at %d", rewindCount, RewindFrames)
	}
	if got := rewindHead - rewindOldest() + 1; got != RewindFrames {
		t.Errorf("the window spans %d frames, want %d", got, RewindFrames)
	}
}

// Seeking is bounded by the recorded history at both ends.
func TestRewindSeekClampsToHistory(t *testing.T) {
	resetRewind()
	recordFrames(100)
	oldest, head := rewindOldest(), rewindHead

	if rewindSeek(-100000); rewindCursor != oldest {
		t.Errorf("rewinding far past the start stopped at %d, want the oldest frame %d", rewindCursor, oldest)
	}
	if moved := rewindSeek(-1); moved {
		t.Errorf("rewinding from the oldest frame reported movement to %d", rewindCursor)
	}
	if slot := rewindSlot(rewindCursor); slot < 0 || slot >= RewindFrames {
		t.Errorf("the playhead maps to slot %d, outside the buffer", slot)
	}

	if rewindSeek(100000); rewindCursor != head {
		t.Errorf("replaying past the live edge stopped at %d, want the head %d", rewindCursor, head)
	}
	if moved := rewindSeek(1); moved {
		t.Errorf("replaying from the live edge reported movement to %d", rewindCursor)
	}

	// An empty recording has nowhere to go.
	resetRewind()
	if moved := rewindSeek(-5); moved {
		t.Errorf("seeking in an empty recording moved the playhead to %d", rewindCursor)
	}
}

// Resuming from a rewound frame drops the future that was recorded after it,
// without touching the other end of the window.
func TestRewindResumeHere(t *testing.T) {
	resetRewind()
	recordFrames(100)
	oldest := rewindOldest()
	rewindSeek(-30)
	resumed := rewindCursor

	rewindResumeHere()

	if rewindHead != resumed {
		t.Errorf("head is %d after resuming at %d, want them equal", rewindHead, resumed)
	}
	if rewindOldest() != oldest {
		t.Errorf("the oldest frame moved to %d, want it left at %d", rewindOldest(), oldest)
	}
	if rewindSeek(1000); rewindCursor != resumed {
		t.Errorf("replaying forward reached %d, want it stopped at the new head %d", rewindCursor, resumed)
	}
	if rewindSeek(-1000); rewindCursor != oldest {
		t.Errorf("rewinding reached %d, want the untouched oldest frame %d", rewindCursor, oldest)
	}
}

// R toggles the mode: one press enters and pauses, the next resumes.
func TestTimeToggle(t *testing.T) {
	ResetTime()
	recordFrames(100)
	if TimeMode != TimeOff {
		t.Fatalf("mode %d before any press, want off", TimeMode)
	}

	UpdateTime(true, 0, 1.0/60)
	if TimeMode != TimePaused {
		t.Errorf("mode %d after the first press, want paused", TimeMode)
	}
	head := rewindHead

	// Held, not pressed again: nothing moves and the mode stays on.
	for i := 0; i < 10; i++ {
		if moved := UpdateTime(false, 0, 1.0/60); moved {
			t.Fatalf("the playhead moved while the mouse was still")
		}
	}
	if TimeMode != TimePaused || rewindCursor != head {
		t.Errorf("holding changed things: mode %d cursor %d, want paused on %d", TimeMode, rewindCursor, head)
	}

	UpdateTime(true, 0, 1.0/60)
	if TimeMode != TimeOff {
		t.Errorf("mode %d after the second press, want off", TimeMode)
	}
}

// Leaving the mode on an earlier frame throws away the future recorded after it.
func TestTimeResumeDropsTheFuture(t *testing.T) {
	ResetTime()
	recordFrames(100)
	UpdateTime(true, 0, 1.0/60)    // enter
	UpdateTime(false, -10, 1.0/60) // rewind one quantum
	rewound := rewindCursor
	if rewound >= rewindHead {
		t.Fatalf("the playhead did not move back: %d of head %d", rewound, rewindHead)
	}
	UpdateTime(true, 0, 1.0/60) // leave

	if rewindHead != rewound {
		t.Errorf("head %d after resuming at %d, want them equal", rewindHead, rewound)
	}
}

// Only the direction of the mouse matters, and only horizontally.
func TestScrubDirection(t *testing.T) {
	for _, c := range []struct {
		name string
		dx   float32
		want int
	}{
		{"moving right", 12, 1},
		{"moving right slowly", 1, 1},
		{"moving left", -12, -1},
		{"moving left slowly", -1, -1},
		{"still", 0, 0},
		{"sub-pixel jitter right", 0.4, 0},
		{"sub-pixel jitter left", -0.4, 0},
		{"straight up or down", 0, 0},
	} {
		if got := scrubDirection(c.dx); got != c.want {
			t.Errorf("%s (dx %v): direction %d, want %d", c.name, c.dx, got, c.want)
		}
	}
}

// drag enters the mode on a full recording, with the playhead at the live edge
// or, from the start, at the oldest frame, and moves the mouse by dx every
// rendered frame for the given seconds at fps. It reports how many frames the
// playhead covered.
func drag(t *testing.T, fromOldest bool, seconds float64, fps int, dx float32) int {
	t.Helper()
	ResetTime()
	recordFrames(RewindFrames)
	UpdateTime(true, 0, 0) // enter the mode
	if fromOldest {
		rewindCursor = rewindOldest()
	}
	from := rewindCursor
	dt := float32(1) / float32(fps)
	for range int(seconds * float64(fps)) {
		UpdateTime(false, dx, dt)
	}
	return rewindCursor - from
}

// covered is what the ramp should cover over n fixed steps from rest.
func covered(n int) (f int) {
	for held := range n {
		f += scrubRate(held)
	}
	return f
}

// The rate starts at ScrubFrames, rises in whole frames and never falls, and
// holds at ScrubFramesMax from the end of the ramp on.
func TestScrubRate(t *testing.T) {
	if got := scrubRate(0); got != ScrubFrames {
		t.Errorf("rate from rest is %d, want %d", got, ScrubFrames)
	}
	for _, held := range []int{ScrubRampSteps, ScrubRampSteps + 1, 10 * ScrubRampSteps} {
		if got := scrubRate(held); got != ScrubFramesMax {
			t.Errorf("rate after %d steps is %d, want %d", held, got, ScrubFramesMax)
		}
	}
	for held := 1; held <= ScrubRampSteps; held++ {
		if scrubRate(held) < scrubRate(held-1) {
			t.Fatalf("the rate fell from %d to %d at step %d", scrubRate(held-1), scrubRate(held), held)
		}
	}
}

// Neither the frame rate nor the direction changes how far a drag goes.
func TestScrubIsTheSameAtAnyFrameRateAndEitherWay(t *testing.T) {
	slow, fast := drag(t, false, 3, 60, -8), drag(t, false, 3, 150, -8)
	if slow != fast {
		t.Errorf("three seconds of rewinding covered %d frames at 60 fps and %d at 150 fps", -slow, -fast)
	}
	if want := -covered(3 * 60); slow != want {
		t.Errorf("three seconds of rewinding covered %d frames, want %d", -slow, -want)
	}
	back, forward := drag(t, false, 2, 60, -8), drag(t, true, 2, 60, 8)
	if -back != forward {
		t.Errorf("two seconds covered %d frames rewinding and %d replaying forward", -back, forward)
	}
	if edge := drag(t, false, 1, 60, 8); edge != 0 {
		t.Errorf("scrubbing forward from the live edge moved %d frames, want none", edge)
	}
}

// A sweep and a drag of the same duration, each from rest, cover the same
// ground: the ramp counts time, not distance.
func TestScrubSpeedOfHandDoesNotMatter(t *testing.T) {
	sweep, slow := drag(t, false, 1, 60, -40), drag(t, false, 1, 60, -2)
	if sweep != slow {
		t.Errorf("a fast sweep covered %d frames and a slow drag %d", -sweep, -slow)
	}
}

// A short drag steps slowly; holding it ramps up to the full rate and reaches
// a minute back in about five seconds.
func TestScrubRamps(t *testing.T) {
	if short := -drag(t, false, 0.25, 60, -8); short >= 60 {
		t.Errorf("a quarter-second drag covered %d frames, want under a second of play", short)
	}
	ResetTime()
	recordFrames(RewindFrames)
	UpdateTime(true, 0, 0)
	perSecond := make([]int, 4)
	for s := range perSecond {
		from := rewindCursor
		for range 60 {
			UpdateTime(false, -8, 1.0/60)
		}
		perSecond[s] = from - rewindCursor
	}
	if !(perSecond[0] < perSecond[1] && perSecond[1] < perSecond[2]) {
		t.Errorf("frames covered per second %v, want them rising through the first two", perSecond)
	}
	if want := ScrubFramesMax * 60; perSecond[2] != want || perSecond[3] != want {
		t.Errorf("frames covered per second %v, want %d once the ramp is done", perSecond, want)
	}
	if far := -drag(t, false, 5, 60, -8); far < RewindFrames-1 {
		t.Errorf("five seconds of dragging reached %d frames back, want the whole minute", far)
	}
}

// scrubFor moves the mouse by dx for n rendered frames at 60 fps and reports
// the frames the playhead covered over them.
func scrubFor(n int, dx float32) int {
	from := rewindCursor
	for range n {
		UpdateTime(false, dx, 1.0/60)
	}
	return rewindCursor - from
}

// Reversing, or stopping for longer than the grace, starts the ramp over; a
// single dropped mouse report does neither. Each is judged over the next ten
// frames rather than one: a rendered frame does not line up with a fixed step,
// so any single frame may cover one step or none.
func TestScrubRampStartsOver(t *testing.T) {
	// three seconds of ramp covers about 2000 frames: from the live edge there
	// is room to rewind that and more, and from the middle room to turn back
	enter := func(middle bool) {
		ResetTime()
		recordFrames(RewindFrames)
		UpdateTime(true, 0, 0)
		if middle {
			rewindCursor = rewindOldest() + RewindFrames/2
		}
	}

	enter(true)
	scrubFor(180, -8)
	if got, max := scrubFor(10, 8), 10*scrubRate(10); got == 0 || got > max {
		t.Errorf("reversing after three seconds covered %d frames in ten, want no more than %d", got, max)
	}

	enter(false)
	scrubFor(180, -8)
	still := rewindCursor
	scrubFor(30, 0) // half a second still
	if rewindCursor != still {
		t.Errorf("time moved while the mouse was still")
	}
	if got, max := -scrubFor(10, -8), 10*scrubRate(10); got == 0 || got > max {
		t.Errorf("rewinding again after a pause covered %d frames in ten, want no more than %d", got, max)
	}

	enter(false)
	scrubFor(180, -8)
	if moved := scrubFor(1, 0); moved != 0 {
		t.Errorf("a dropped report moved time %d frames", moved)
	}
	if got, min := -scrubFor(10, -8), 9*ScrubFramesMax; got < min {
		t.Errorf("after a dropped report the drag covered %d frames in ten, want it to carry on at %d a step", got, ScrubFramesMax)
	}
}

// Sensor noise neither moves time nor winds the ramp up.
func TestScrubJitterDoesNothing(t *testing.T) {
	ResetTime()
	recordFrames(RewindFrames)
	UpdateTime(true, 0, 0)
	from := rewindCursor
	for i := range 180 {
		dx := float32(0.3)
		if i%2 == 0 {
			dx = -0.3
		}
		UpdateTime(false, dx, 1.0/60)
	}
	if rewindCursor != from || scrubHeld != 0 {
		t.Errorf("jitter moved time %d frames and wound the ramp to %d", rewindCursor-from, scrubHeld)
	}
}

// Stopping the mouse stops time where it is.
func TestScrubStopsWithTheMouse(t *testing.T) {
	ResetTime()
	recordFrames(RewindFrames)
	UpdateTime(true, 0, 0)
	for i := 0; i < 10; i++ {
		UpdateTime(false, -8, 1.0/60)
	}
	stopped := rewindCursor
	for i := 0; i < 10; i++ {
		if moved := UpdateTime(false, 0, 1.0/60); moved {
			t.Fatalf("the playhead moved after the mouse stopped")
		}
	}
	if rewindCursor != stopped || TimeMode != TimePaused {
		t.Errorf("cursor %d mode %d after stopping, want %d and paused", rewindCursor, TimeMode, stopped)
	}
}

// One rule answers "is a paused mode holding the match?" for every pause path,
// so no path can be left out of a gate the way time manipulation once was.
func TestMatchHeld(t *testing.T) {
	defer func() { Editing = false; ResetTime() }()
	ResetTime()
	Editing = false
	if matchHeld() {
		t.Errorf("the match is held during normal play")
	}
	Editing = true
	if !matchHeld() {
		t.Errorf("the match is not held while edit mode is on")
	}
	Editing = false
	for _, mode := range []int{TimePaused, TimeScrubbing} {
		TimeMode = mode
		if !matchHeld() {
			t.Errorf("the match is not held while time manipulation is in mode %d", mode)
		}
	}
	TimeMode = TimeOff
	if matchHeld() {
		t.Errorf("the match is still held once both modes are off")
	}
}

// A click made while time manipulation holds the match must not fire when the
// match resumes: no fixed step runs during the pause, so nothing consumes the
// latch unless leaving the mode drops it.
func TestLeavingTimeManipulationDropsLatchedInput(t *testing.T) {
	defer func() { ResetTime(); clearLatchedInput() }()
	ResetTime()
	recordFrames(10)

	UpdateTime(true, 0, 1.0/60) // enter
	if !matchHeld() {
		t.Fatalf("time manipulation is on but the match is not held")
	}
	pressedAttack, pressedJump = true, true // clicked while the match is held

	UpdateTime(true, 0, 1.0/60) // leave
	if pressedAttack || pressedJump {
		t.Errorf("leaving time manipulation kept latched input: attack %t jump %t", pressedAttack, pressedJump)
	}
	if matchHeld() {
		t.Errorf("the match is still held after leaving time manipulation")
	}
}
