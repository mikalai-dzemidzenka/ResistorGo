## Why

Time manipulation promises the player can pick the match back up from any
moment they find, but only where the fighters stood and how they were posed is
recorded. Leaving the mode on a past frame puts everyone back in place and
leaves everything else as it was at the live edge: structure already lost stays
lost, a bot killed after that frame is still dead, a swing already thrown is
still on its way. Now that a fallen bot stays lying where it fell in single
player, this is plain to see: scrub back to a bot still standing, resume, and
the body slides to where it stood and falls over again.

A fighter that came into the world partway through the recording also has
nothing recorded for the frames before it did. Scrubbing back past a respawn
draws the new fighter at the origin with no sprite, and resuming there puts it
at the origin, while the fighter it replaced is missing from those frames
entirely.

## What Changes

- Every recorded frame holds the whole single-player match as it stood after
  that step: every fighter in the world with its full simulation state, the
  fighters waiting to respawn and those waiting to be taken out, and the
  world's tick.
- Scrubbing shows the fighters that were in the world on the frame shown, and
  only those: a fighter already removed on that frame is back, and one not yet
  spawned is absent.
- Leaving the mode on a past frame resumes the match from that frame
  completely: structure, energy, stuns, deaths, bodies, attacks under way,
  queued and buffered input, the bots' minds, whose turn it is at the front of
  a fight, and who is waiting to respawn and when.
- **BREAKING** (to a stated rule, not an interface): the bot's mind and the
  places at a fight, documented as never rewound, are rewound in single
  player. They stay unreplicated; networked play is untouched.
- The per-actor pose ring (`ActorRewindFrame`, `Actor.RewindData`) is replaced
  by one ring of world frames, which also removes the unrecorded frames of a
  fighter spawned mid-recording.

Out of scope: networked play, where time manipulation stays unavailable; the
length of the recording (256 steps); the restart key, which still builds a
fresh match with `newOfflineWorld` rather than seeking to a frame that may have
fallen out of the recording. Randomness is not recorded: a match resumed from
the past takes the same state down a new future.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities
- `time-manipulation`: leaving the mode on a past frame restores the whole
  match, not only positions; the frames shown hold exactly the fighters that
  were in the world then.

## Impact

`rewind.go` gains the world-frame ring and its capture and restore, replacing
`ActorRewindFrame`. `actor.go` loses `RewindData`, `recordRewind` and
`restoreRewind`; `spriteDraw` reads the pose from the restored actor rather
than from a separate recording. `world.go` exposes what a frame needs of it
(tick, next ID, pending spawns, dying list, local fighter). `game.go`'s
`stepOffline` records the world once per step, and its time branch restores the
world instead of each actor. The comments on `Actor.Engaged` and `Actor.Bot`
that say they are never rewound are updated.

Anything outside the world that holds a fighter pointer (`cursorOwner`, the
draw list, the camera's target) must follow the restored fighters rather than
the replaced ones. Tests: `rewind_test.go` and the tests that call
`rewindRecord` directly move to the new recording.
