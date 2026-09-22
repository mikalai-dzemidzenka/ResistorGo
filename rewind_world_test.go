package resistor

import (
	"slices"
	"testing"
)

// rewindMatch is offlineMatch with an empty recording, left empty afterwards.
func rewindMatch(t *testing.T) *Actor {
	t.Helper()
	pl := offlineMatch(t)
	ResetTime()
	t.Cleanup(ResetTime)
	return pl
}

// resumeAt seeks to frame and leaves time manipulation there, as the player
// scrubbing to it and pressing T would.
func resumeAt(w *World, frame int) {
	rewindCursor = frame
	restoreFrame(w)
	rewindResumeHere()
}

// recStep runs one step of runFight and records it, as stepOffline would.
func recStep(w *World, p *Actor) {
	runFight(w, 1, p)
	rewindRecord()
	captureFrame(w)
}

// asRestored is a recorded value as a restore puts it back: at rest, with its
// events already played.
func asRestored(v Actor) Actor {
	v.PrevPos, v.InterpolatedPos, v.Events = v.Pos, v.Pos, 0
	return v
}

// Once the ring has gone round, recording a step reuses what the last lap left
// in the slot and allocates nothing.
func TestCaptureDoesNotAllocate(t *testing.T) {
	rewindMatch(t)
	for range 3 {
		CL.Spawn(CL.spawnPos(), false)
	}
	for range RewindFrames + 10 {
		rewindRecord()
		captureFrame(CL)
	}
	if n := testing.AllocsPerRun(100, func() {
		rewindRecord()
		captureFrame(CL)
	}); n != 0 {
		t.Errorf("%v allocations a step", n)
	}
}

// A restored frame is every fighter that was in the world, exactly as it was,
// and nobody else - whatever happened in between.
func TestRestoreIsTheRecordedMatch(t *testing.T) {
	pl := rewindMatch(t)
	bot := CL.Spawn(CL.spawnPos(), false)
	ticks(10)
	mark := rewindHead
	saved := map[*Actor]Actor{pl: *pl, bot: *bot}

	bot.takeHit(pl)
	CL.Kill(pl)
	ticks(RespawnTicks + 10)
	if CL.Local == pl || CL.Local == nil {
		t.Fatalf("the player did not respawn: nothing to undo")
	}
	respawned := CL.Local

	resumeAt(CL, mark)
	if !slices.Equal(CL.Actors, []*Actor{pl, bot}) {
		t.Fatalf("world holds %v, want the player and the bot as they were", CL.Actors)
	}
	for a, v := range saved {
		if *a != asRestored(v) {
			t.Errorf("fighter %d is not as it was recorded", a.ID)
		}
	}
	if slices.Contains(CL.Actors, respawned) {
		t.Error("the fighter that respawned later is still in the world")
	}
	if CL.Local != pl {
		t.Error("the player does not control the fighter that had not died yet")
	}
}

// Resuming plays on from a copy: the recorded frame is not the live match.
func TestResumedMatchDoesNotEditTheRecording(t *testing.T) {
	pl := rewindMatch(t)
	bot := CL.Spawn(CL.spawnPos(), false)
	ticks(10)
	mark := rewindHead
	f := &rewindFrames[rewindSlot(mark)]
	before := slices.Clone(f.actors)

	resumeAt(CL, mark)
	bot.takeHit(pl)
	CL.Kill(pl)
	ticks(60)

	if !slices.Equal(f.actors, before) {
		t.Error("playing on from a frame changed what the frame holds")
	}
}

// A restart holds on to nothing of the match it threw away.
func TestResetTimeLetsGoOfTheMatch(t *testing.T) {
	rewindMatch(t)
	CL.Spawn(CL.spawnPos(), false)
	ticks(30)
	ResetTime()
	for i, f := range rewindFrames {
		if f.local != nil || len(f.actors) != 0 || cap(f.actors) != 0 || cap(f.dying) != 0 {
			t.Fatalf("slot %d still refers to the old match", i)
		}
	}
}

// The newest frame is the match on screen: recorded after the step, not before.
func TestNewestFrameIsTheLiveMatch(t *testing.T) {
	rewindMatch(t)
	CL.Spawn(CL.spawnPos(), false)
	ticks(20)
	f := rewindFrames[rewindSlot(rewindHead)]
	if f.tick != CL.Tick || len(f.actors) != len(CL.Actors) {
		t.Fatalf("newest frame is tick %d with %d fighters, live is tick %d with %d", f.tick, len(f.actors), CL.Tick, len(CL.Actors))
	}
	for i, r := range f.actors {
		if r.a != CL.Actors[i] || r.v != *CL.Actors[i] {
			t.Errorf("fighter %d on the newest frame is not the live one", r.a.ID)
		}
	}
}

// Beat a bot, rewind to before, resume: it gets back up, and stays up.
func TestResumingBeforeAKillUndoesIt(t *testing.T) {
	pl := rewindMatch(t)
	CL.BotsStayDead = true
	bot := CL.Spawn(CL.spawnPos(), false)
	ticks(10)
	mark, structure := rewindHead, bot.Structure

	bot.enterDeathState(pl)
	ticks(200)
	if bot.State != StateDeath {
		t.Fatalf("the bot is not lying down: nothing to undo")
	}

	resumeAt(CL, mark)
	if bot.State == StateDeath || bot.Structure != structure {
		t.Fatalf("resumed bot is in state %d with structure %v, want alive with %v", bot.State, bot.Structure, structure)
	}
	ticks(200)
	if bot.State == StateDeath {
		t.Error("the bot fell again with nobody fighting it")
	}
}

// Die, respawn, rewind to before the death: the old fighter is back in
// the player's hands and no second one is on its way.
func TestResumingBeforeThePlayersDeathUndoesIt(t *testing.T) {
	pl := rewindMatch(t)
	ticks(10)
	mark := rewindHead

	CL.Kill(pl)
	ticks(RespawnTicks + 10)
	respawned := CL.Local
	if respawned == nil || respawned == pl {
		t.Fatalf("the player did not respawn: nothing to undo")
	}

	resumeAt(CL, mark)
	if CL.Local != pl || !slices.Contains(CL.Actors, pl) || pl.Dying {
		t.Fatal("the player is not back on the fighter that had not died")
	}
	if slices.Contains(CL.Actors, respawned) {
		t.Error("the respawned fighter is still in the world")
	}
	if len(CL.pending) != 0 || len(CL.dying) != 0 {
		t.Errorf("%d respawns and %d removals still waiting", len(CL.pending), len(CL.dying))
	}
	ticks(RespawnTicks + 10)
	if CL.Local != pl || len(CL.Actors) != 1 {
		t.Error("a respawn from the discarded future came through")
	}
}

// A frame from before a fighter came into the world does not have it.
func TestAFighterIsAbsentBeforeItExisted(t *testing.T) {
	rewindMatch(t)
	ticks(10)
	mark := rewindHead
	late := CL.Spawn(CL.spawnPos(), false)
	ticks(10)

	rewindCursor = mark
	restoreFrame(CL)
	if slices.Contains(CL.Actors, late) {
		t.Error("a fighter spawned later is on a frame from before")
	}
}

// Rewind to before a hit and resume: none of the hit is left.
func TestResumingBeforeAHitUndoesTheDamage(t *testing.T) {
	pl := rewindMatch(t)
	bot := CL.Spawn(CL.spawnPos(), false)
	ticks(10)
	mark := rewindHead
	was := *bot

	bot.takeHit(pl)
	ticks(2)
	resumeAt(CL, mark)
	if bot.Structure != was.Structure || bot.Energy != was.Energy || bot.HitCount != was.HitCount || bot.State != was.State {
		t.Errorf("after resuming: structure %v energy %d hits %d state %d, want %v %d %d %d",
			bot.Structure, bot.Energy, bot.HitCount, bot.State, was.Structure, was.Energy, was.HitCount, was.State)
	}
}

// Resume in the middle of a swing and it goes on from there: the same frames,
// landing on the same tick.
func TestASwingInProgressResumesWhereItWas(t *testing.T) {
	ResetTime()
	t.Cleanup(ResetTime)
	w := NewWorld(true)
	p, bots := lineUp(w, 1)
	b := bots[0]
	for range 600 {
		recStep(w, p)
		if b.State == StateAttack && b.AnimCurrentFrame > 0 {
			break
		}
	}
	if b.State != StateAttack {
		t.Fatal("the bot never swung: nothing to resume")
	}
	mark := rewindHead

	type seen struct {
		tick   uint32
		frame  uint8
		landed bool
		state  ActorState
	}
	play := func() []seen {
		var got []seen
		for b.State == StateAttack && len(got) < 120 {
			recStep(w, p)
			got = append(got, seen{w.Tick, b.AnimCurrentFrame, b.SwingHit, p.State})
		}
		return got
	}
	first := play()
	resumeAt(w, mark)
	if again := play(); !slices.Equal(first, again) {
		t.Errorf("the swing went\n%v\nthe first time and\n%v\nafter resuming", first, again)
	}
}

// The line at a fight comes back as it was on the frame: the same two at the
// front and the same one waiting, however it had changed since.
func TestTheLineAtAFightIsAsItWas(t *testing.T) {
	ResetTime()
	t.Cleanup(ResetTime)
	w := NewWorld(true)
	w.BotsStayDead = true
	p, bots := lineUp(w, 3)
	for range 240 {
		recStep(w, p)
	}
	front := engaged(p)
	if len(front) != 2 {
		t.Fatalf("%d bots at the front, want 2", len(front))
	}
	mark := rewindHead
	type mind struct {
		target *Actor
		side   int8
	}
	minds := map[*Actor]mind{}
	for _, b := range bots {
		minds[b] = mind{b.Bot.target, b.Bot.side}
	}
	places := p.Engaged

	front[0].Structure = HitStructure
	front[0].enterDeathState(p)
	for range 240 {
		recStep(w, p)
	}
	if p.Engaged == places {
		t.Fatal("the line did not change: nothing to restore")
	}

	resumeAt(w, mark)
	if p.Engaged != places {
		t.Errorf("places at the fighter are %v, want %v", p.Engaged, places)
	}
	for _, b := range bots {
		if got := (mind{b.Bot.target, b.Bot.side}); got != minds[b] {
			t.Errorf("bot %d is fighting %v on side %d, want %v on %d", b.ID, got.target, got.side, minds[b].target, minds[b].side)
		}
	}
}
