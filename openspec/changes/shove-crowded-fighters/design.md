## Context

See proposal.md - Why.

All offsets are `dx = other.Pos.X - self.Pos.X`. A fighter's body is the
collision box `[P+14, P+28]`; with the reach corrected by `symmetric-swing-reach`
a swing (dash included) covers 35px in front of the body's centre, the same both
ways:

```
  overlap              -14 < dx < 14
  reach facing right    -3 < dx < 35
  reach facing left    -35 < dx < 3

  crowding = overlap and not reach
             facing right:  dx in [-13, -3]
             facing left:   dx in [  3, 13]
```

The window is always behind the fighter, never in front, and never at `dx = 0`.
It is also one-sided: a fighter at `dx = -8` behind a right-facing one has that
one at `+8` in front of it, inside its own reach whenever it faces it.

What exists today: the state enum ends in a `stateCount` sentinel and the
snapshot decoder is bounded by it, so an appended state survives the wire.
`inSwingRange` in `bot.go` computes reach inline. Every movement test runs on
`floor()`, a level with no wall. A player standing still faces the mouse every
tick (`Actor.Update`, the `StateIdle` branch); a bot faces its target when in
range and its last sighting otherwise.

A first implementation was built against the old premise and withdrawn; it is
saved, with its `push_test.go`, in `recovered.md`. Parts carry over unchanged;
the decisions below mark what does not.

## Goals / Non-Goals

**Goals:**

- A shove that only answers the one-sided position, and leaves the two facing one
  another at arm's length.
- One expression of "can this fighter reach that one", shared by the shove and
  the bot.
- No shove that can repeat itself, without a guard written for the purpose.

**Non-Goals:**

- A deliberate, player-pressed shove. It is automatic.
- Any change to reach. That is `symmetric-swing-reach`, committed.
- Resolving spawn piles: more fighters than spawn points still stack at the
  origin, and now shove on arrival.

## Decisions

### Only from standing

The trigger is `StateIdle` on the ground, not guarding everything, and nothing
else. The withdrawn version also shoved from `StateRun`, and in the bot course
test it fired on 142 of 500 ticks: whoever is being followed always has its
follower behind it. A running fighter has its own answer - it is already moving
away - so standing is the only state in which being on someone's back is a
position rather than a pursuit.

Alternatives: shoving from any grounded state (the chase problem, and it would
cut swings short); requiring the crowd to persist N ticks (a tuning knob with no
evidence it is needed yet - can be added later without changing the state or the
art).

### Turn toward the crowder, then push it away

`enterPushState(other)` faces the crowder, then the shove moves it away from the
shover:

```
  before (A standing, facing right)     after (A facing left)

      B  A -->                            B          <-- A
    [==][==]                            [==]  . . . [==]
    dx in [-13, -3]                     dx in [-29, -19]
    B can hit A; A cannot hit B         apart, face to face, both in reach
```

The withdrawn version first turned toward the other and then dropped the turn,
because at `dx = 0` "which side the other is on is a pixel of noise" and a bot
turned away from its target as often as toward it. The window no longer contains
`dx = 0`; the nearest a crowder can be is 3px to one side, so the side is always
unambiguous and the turn is always toward the fighter the bot is fighting.

The alternative - shove the way the fighter already faces, throwing the crowder
through it to its front - lands inside reach too, but `push.png` shows both arms
pushing forward, and the crowder passing through the body to meet them reads as a
teleport.

For a player, the facing the shove sets holds while it plays; on returning to
standing the mouse decides again, as it always does from standing. Nothing is
lost: the crowder is no longer overlapping, so wherever the player then faces
cannot set off another shove.

### `PushDistance` = 16, bounded both ways

The crowder starts 3 to 13px from the shover and is moved `PushDistance` further:

```
  must clear the overlap:   3 + PD >= 14   ->  PD >= 11
  must stay in reach:      13 + PD <  35   ->  PD <= 21

   11 ............. 16 ............. 21
```

The withdrawn value, 20, satisfies both but lands a deep crowder at 33 of 35px -
two pixels from out of reach, which a later change to `AttackDash` or a box width
would silently close. 16 sits in the middle and lands the crowder 19 to 29px out.
Both bounds are asserted by a test sweeping every offset in the window, not by
the comment beside the constant.

### No repeat guard: the turn is the guard

A wall can cut the throw short, leaving the crowder where it was. The withdrawn
design needed an explicit "do not re-enter `StatePush`" guard for that. With the
turn, the shover ends facing the crowder, and anything overlapping in front is
inside reach, so `crowding()` is false and no second shove begins:

```
  |B [==] <-- A [==]     wall stops B; A faces left; B at dx in [-13, -3]
  |                      is inside the left-facing reach (-35, 3)
```

This is proven by a test against a wall, which needs a `walls()` helper beside
`floor()`.

### Displacement first, then speed, then the wall

`shove` moves `Pos.X` directly, because a fighter mid-swing clamps its own speed
to the lunge and a speed-only shove would leave it standing still. It sets the
crowder's speed *before* `blockSides`, so a throw a wall cuts short keeps the
clamp's removal of wall-ward speed rather than overwriting it and pressing the
fighter into the wall.

### Reach is asked in one place

`swingReachFacing(t, o)` is lifted out of `inSwingRange` with its block-all
fallback (measure the swing the fighter would throw the moment it stopped
guarding). `inSwingRange` becomes a one-line caller; `crowding` asks it with the
fighter's own facing.

### Carried over unchanged from the withdrawn version

`StatePush` and `push.png` appended; `PushSeconds` 0.35 over `PushFrames` 7,
landing on `PushFrame` 3; the `PushDone` once-latch; the dead neither shove nor
are shoved; the displacement lands only where `W.Authority`; `StatePush` excluded
from `canMove`; `jump_test.go`'s fighters spaced apart.

## Risks / Trade-offs

- **[A client predicts a turn the server does not make]** -> The crowding check
  is not authority-gated, only the displacement is, so a predicting client turns
  and plays the shove on its own reading of a lagged remote fighter. The turn
  makes a misprediction more visible than before: the local fighter flips, then
  is corrected. Measured in play under latency simulation (tasks); if it shows,
  the check moves behind the authority and the client plays whatever `StatePush`
  the snapshot carries, at the cost of a round trip.
- **[Back to back, both are crowded]** -> Two fighters facing away from one
  another can each be behind the other; both shove on the same tick and turn to
  face. The one earlier in `W.Actors` lands first, and the second's overlap
  re-check then fails, so exactly one displacement happens. Consistent, but slice
  order picks who moves.
- **[The one behind hits during the wind-up]** -> By design the shove is
  interruptible, so a fighter camping someone's back with swings keeps them in
  `StateAttacked` and never lets them shove. That is the reward for getting
  there; turning around by mouse is still the answer between hits.
- **[Tests assumed a quiet world]** -> Anything that stands two fighters within
  14px of each other now has a shove in it. Swept by running the suite
  repeatedly, not by inspection.

## Migration Plan

`StatePush` and `push.png` are appended, so no existing state or animation id
moves. Client and server ship together: an old client decoding a snapshot with
`StatePush` would reject it as out of range. Rollback is reverting the change.
