## Context

See proposal.md for why. Today `stepOffline` calls `rewindRecord()` and then
each actor's `recordRewind()` at the start of every fixed step, writing a pose
(`Pos`, `Speed`, `Orientation`, `Anim`, `AnimFrame`) into a 256-slot array
embedded in `Actor` (`RewindData`). `restoreRewind` puts back position, motion
and facing only; `spriteDraw` draws the recorded pose while the mode is on.
The ring's bookkeeping (`rewindHead`, `rewindCount`, `rewindCursor`,
`rewindSeek`, `rewindResumeHere`) is independent of what a slot holds, and is
kept.

What a single-player match is, at the level a frame must capture:

```
  World: Actors []*Actor, Local *Actor, Tick, nextID,
         pending []pendingSpawn, dying []*Actor
         (Authority is fixed; History is nil offline)

  Actor: plain values, plus pointers
         W *World, Attacker *Actor, PushStuck *Actor,
         Engaged [2]*Actor, Bot.target *Actor, AnimTop *string
         no slices or maps apart from RewindData, which goes
```

An `Actor` object is never reused once it leaves the world: `World.Remove`
drops it from `Actors` and clears the pointers that name it, a respawn builds a
new object with a new ID. Outside the world, `cursorOwner` and `drawList` in
`game.go` hold fighter pointers; the camera follows a position, not a pointer.

## Goals / Non-Goals

**Goals:**
- Restoring a frame is a copy, not a reconstruction: no per-field list of what
  to put back that a new field can silently miss.
- No pointer relinking, so no class of bug where a restored fighter points at
  an object from the other timeline.
- Frames and the live match never share mutable memory.

**Non-Goals:**
- Recording anything outside the world: camera, particles, sound, the HUD's
  flashes, `GameTimeSec`, `PhysicsTime`, the random generator.
- Replaying the discarded future deterministically.
- Any use of this in networked play.

## Decisions

### A frame keeps the actor objects themselves, and a copy of each one's value

```
  worldFrame
  +---------------------------------------------+
  | tick, nextID                                |
  | local   *Actor                              |
  | pending []pendingSpawn   (own copy)         |
  | dying   []*Actor         (own copy)         |
  | actors  []actorFrame                        |
  |           { a *Actor ; v Actor }            |
  +---------------------------------------------+

  capture:  for each a in w.Actors: {a, *a}
  restore:  for each f: *f.a = f.v ; w.Actors = [f.a ...]
```

Because an object keeps its identity for its whole life and is never reused,
every pointer in a recorded value (`Engaged`, `Attacker`, `PushStuck`,
`Bot.target`, `Local`, the entries of `dying`) names an object that the same
frame also records, or one that had already left the world by then and is
held exactly as the live match held it on that frame (`World.Remove` clears
most such pointers; `PushStuck` it leaves, harmlessly, as it does live). Restoring every recorded object's value
therefore brings back a consistent graph with nothing to relink. A fighter
removed after the frame is reinstated as the very object it was, and one
spawned after the frame is simply not in the restored `Actors`.

Alternative considered: store values with pointers turned into IDs, and on
restore rebuild fresh objects and relink by ID. It needs a hand-written
mapping for every pointer field, including ones added later, and it gives
fighters new identities on every restore, which `cursorOwner` and anything
else comparing pointers would have to follow. Keeping the objects avoids both.

The cost is that a frame holds removed fighters alive for as long as the frame
exists: at most 256 frames, and only in single player.

### `RewindData` leaves `Actor`

A value copy of `Actor` with a 256-slot array inside it would copy the array on
every capture and nest recordings inside recordings. The ring moves to
`rewind.go` as `[RewindFrames]worldFrame`, and `ActorRewindFrame`,
`recordRewind` and `restoreRewind` are removed. With the whole value restored,
drawing a recorded frame is drawing the live actors: `spriteDraw` and `Draw`
lose their `TimeShowsHistory` branches for the pose.

Slots are reused in place: a capture overwrites the slot's `actors` and
`pending`/`dying` slices down to length zero and appends, so steady play does
not allocate.

### Record after the step, not before

Today the frame is opened before the actors update, so the newest frame is one
step behind what is on screen. The frame is captured at the end of each fixed
step instead: frame N is the match after tick N, the live edge is exactly the
screen, and resuming from frame N continues with tick N+1 (`updateRespawns`
first, as always). `rewindRecord` keeps opening the slot; the capture fills it
after the actors have run.

### Restore touches what rendering reads, and clears presentation events

After copying the values back, each restored actor's `PrevPos` and
`InterpolatedPos` are set to its `Pos` (a recorded frame is not in motion),
and `Events` is cleared so a hit or swing recorded on that frame does not play
its sound or flash again. `Accumulator` is zeroed on leaving the mode, as it
already is, so resuming does not catch up.

### The restart keeps its own path

`restartOffline` still calls `ResetTime` and builds a new world. `ResetTime`
also empties the ring's slots, so a restart drops the old match's objects and
the old `*World` rather than holding them for 256 frames.

### The rule "never rewound" becomes "rewound in single player"

The comments on `Actor.Engaged` and `Actor.Bot` are rewritten: neither is
replicated, and both are recorded and restored with the rest of the fighter by
time manipulation. Nothing about networked play changes.

## Risks / Trade-offs

- [A field holds a slice, map or pointer to shared mutable data added later]
  → a frame and the live match would share it and a resumed match would edit
  history. Mitigation: a test that captures a frame, mutates the live match
  heavily, restores, and compares every actor to its captured value; and a
  comment on `Actor` that its value is copied into the recording.
- [Something outside the world holds a fighter that is not in the restored
  frame] → it keeps a fighter that has left the world. `cursorOwner` is already
  compared against the local fighter every frame and re-taken; `drawList` is
  rebuilt every frame. Anything new of this kind has the same obligation.
- [Frame size] → an `Actor` without `RewindData` is a few hundred bytes; a
  frame with the player and three bots is about 2 KB, the ring about 0.5 MB,
  against about 10 KB per actor for the pose ring today.
- [Capture cost per step] → one struct copy per actor per step; negligible
  next to the step itself.
- [The rewound future differs from the discarded one] → by design; the bots
  draw new random numbers. The spec says so.

## Open Questions

- The structure and energy bars are hidden while a recorded frame is shown,
  because until now they would have shown the live edge's values. They are now
  truthful and could stay up while scrubbing; that is a presentation choice
  that can be made during implementation.
