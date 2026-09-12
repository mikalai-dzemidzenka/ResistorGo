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

// scrubbed runs one second of mouse movement as n rendered frames and reports
// how far the playhead travelled.
func scrubbed(t *testing.T, frames int, dx float32) int {
	t.Helper()
	ResetTime()
	recordFrames(RewindFrames)
	UpdateTime(true, 0, 0) // enter the mode
	from := rewindCursor
	dt := float32(1.0) / float32(frames)
	for i := 0; i < frames; i++ {
		UpdateTime(false, dx, dt)
	}
	return rewindCursor - from
}

// The scrub rate is fixed: it does not depend on the frame rate, and it is the
// same in both directions.
func TestScrubRateIsFixed(t *testing.T) {
	slow, fast := scrubbed(t, 60, -8), scrubbed(t, 150, -8)
	if slow != fast {
		t.Errorf("one second of rewinding covered %d frames at 60 fps and %d at 150 fps", -slow, -fast)
	}
	if want := -60 * ScrubFrames; slow != want {
		t.Errorf("one second of rewinding covered %d frames, want %d", slow, want)
	}
	if forward := scrubbed(t, 60, 8); forward != 0 {
		t.Errorf("scrubbing forward from the live edge moved %d frames, want none", forward)
	}
}

// A sweep and a drag of the same duration cover the same ground.
func TestScrubSpeedOfHandDoesNotMatter(t *testing.T) {
	sweep, drag := scrubbed(t, 60, -40), scrubbed(t, 60, -2)
	if sweep != drag {
		t.Errorf("a fast sweep covered %d frames and a slow drag %d", -sweep, -drag)
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
