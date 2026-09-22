> Builds on the `offline-respawn-handover` work (bodies that stay down,
> `newOfflineWorld`, T for time and R for restart); commit that first.

## 1. The world frame

- [x] 1.1 In `rewind.go`, add `actorFrame {a *Actor; v Actor}` and `worldFrame`
      (tick, nextID, local, pending, dying, actors) and a
      `[RewindFrames]worldFrame` ring; verify it compiles alongside the old pose
      ring
- [x] 1.2 Add `captureFrame(w *World)`, writing the slot at `rewindHead` by
      truncating and appending into the slot's own slices; verify with a test
      that capturing 300 steps of a four-fighter match allocates nothing after
      the ring has wrapped once (`testing.AllocsPerRun`)
- [x] 1.3 Add `restoreFrame(w *World)`, copying each value back into its object,
      setting `w.Actors`, `Local`, `Tick`, `nextID`, `pending` and `dying` from
      copies, setting `PrevPos` and `InterpolatedPos` to `Pos` and clearing
      `Events`; verify with a test that captures a frame, runs the match on
      (hits, a death, a respawn), restores, and finds every actor equal to its
      captured value and the respawned fighter gone
- [x] 1.4 Verify isolation: after a restore, run the match on and check that the
      recorded frame's values have not changed
- [x] 1.5 Make `ResetTime` clear every slot's slices and pointers; verify with a
      test that after `ResetTime` no slot refers to an actor

## 2. Wire it into the match

- [x] 2.1 In `stepOffline`, open the frame with `rewindRecord()` as now but call
      `captureFrame(CL)` at the end of the step, after the actors have run;
      verify with a test that the newest frame equals the live match
- [x] 2.2 In `runLevel`'s time branch, call `restoreFrame(CL)` when the playhead
      moves, in place of each actor's `restoreRewind`; verify the existing
      `rewind_test.go` scrub tests pass
- [x] 2.3 Remove `ActorRewindFrame`, `Actor.RewindData`, `recordRewind` and
      `restoreRewind`; drop the recorded-pose branches in `spriteDraw` and
      `Draw`; move `attack_outline_test.go` and `structure_bar_test.go` off
      `rewindRecord` alone onto a captured frame; verify `go test ./...` passes
      (they turned out not to need it: they only check that the bars and the
      outline hide while a recorded frame is shown, which `TimeShowsHistory`
      still decides from the ring's count alone)
- [x] 2.4 Rewrite the comments on `Actor.Engaged` and `Actor.Bot` (neither
      replicated; both recorded and restored by time manipulation), and add a
      note on `Actor` that its value is copied into the recording, so it must
      hold no slice, map or shared mutable pointer; verify by reading

## 3. The spec's scenarios as tests

- [x] 3.1 Resuming before a kill undoes it: beat a bot in a world with
      `BotsStayDead`, seek back to a frame where it stands, resume, run 200
      ticks with no one attacking; verify it is alive with the recorded
      structure
- [x] 3.2 Resuming before the player's death undoes it: the player dies and
      respawns, seek to before the death, resume; verify `CL.Local` is the
      original fighter, the respawned one is not in the world, and nothing is
      pending
- [x] 3.3 A fighter is absent before it existed: spawn a fighter mid-recording,
      seek to before; verify it is not in `CL.Actors`
- [x] 3.4 Resuming before a hit: verify structure, energy, hit count and state
      match the frame
- [x] 3.5 A swing in progress resumes: resume mid-attack; verify the attack
      animation continues from the recorded frame and the hit lands on the tick
      it would have
- [x] 3.6 The line at a fight is as it was: three bots on the player, seek to a
      frame with two engaged and one waiting after the line has changed; verify
      `Engaged` and each bot's target and side match the frame

## 4. In the game

- [ ] 4.1 Build with `build.ps1`; in single player, beat a bot, press T, scrub
      back to before its last hit, press T; verify it stands and fights on
      without teleporting or falling
- [ ] 4.2 Die, respawn, rewind past the death and resume; verify you play the
      old fighter, the camera follows it, and the energy cursor is its own
- [ ] 4.3 Scrub back and forth across a kill and a respawn; verify fighters
      appear and vanish on the frames they entered and left, and no hit sound
      or flash replays when scrubbing over a hit
- [ ] 4.4 Press R while the mode is on and after a resume; verify the level
      restarts cleanly with no history
