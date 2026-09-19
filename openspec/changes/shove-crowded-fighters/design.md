## Context

See proposal.md - Why.

All offsets are `dx = other.Pos.X - self.Pos.X`. A fighter's body is the
collision box `[P+14, P+28]`; with the reach corrected by `symmetric-swing-reach`
a swing (dash included) covers 35px in front of the body's centre, the same both
ways:

```
  overlap              -14 < dx < 14      <- what sets a shove off
  reach facing right    -3 < dx < 35
  reach facing left    -35 < dx < 3
```

What exists today: the state enum ends in a `stateCount` sentinel and the
snapshot decoder is bounded by it, so an appended state survives the wire.
`inSwingRange` in `bot.go` computes reach inline. Every movement test runs on
`floor()`, a level with no wall. A player standing still faces the mouse every
tick (`Actor.Update`, the `StateIdle` branch); a bot faces its target when in
range and its last sighting otherwise. `StateIdle` covers walking as well as
standing - anything below `RunReach`.

A first implementation was built against an older premise and withdrawn; it is
saved, with its `push_test.go`, in `recovered.md`.

## Goals / Non-Goals

**Goals:**

- No two fighters on the ground left standing in one another.
- A shove that leaves the two facing one another at arm's length, and finishes
  once started.
- No shove that can repeat itself against the same fighter from the same spot.

**Non-Goals:**

- A deliberate, player-pressed shove. It is automatic.
- Any change to reach. That is `symmetric-swing-reach`, committed.
- Resolving spawn piles: more fighters than spawn points still stack at the
  origin, and now shove on arrival.

## Decisions

### Any overlap, from anyone on its feet and free

The trigger is any overlap of collision boxes, in front, behind or on the same
column, thrown by a fighter in `StateIdle` or `StateRun` - standing, walking or
running - on the ground, not guarding everything and not flinching from a hit.

How it got here, since each step was tried in play or in a test:

- Behind only, standing only. The first version answered only the one-sided
  position - a fighter on another's back, which can hit it and cannot be hit
  back. In play, overlaps in front were left standing, which read as a glitch.
- Any overlap was chosen instead. Running was kept in: an overlap during a chase
  is still two fighters in one another. The cost, measured in the withdrawn
  version, is that a chase shoves often - whoever is caught up with shoves its
  pursuer. The stun it throws is also what keeps the one caught up with from
  being shoved straight back.
- Flinching is not free. A flinch leaves the fighter in control, but with any
  overlap counting, a flinched shover started a fresh shove on the next tick and
  the hit that interrupted it meant nothing. For the 12 ticks of a flinch it
  does not shove.
- A swing, a block or a stun is finished first; they already own the fighter,
  and cutting them short was the most aggressive option offered and not chosen.

Who shoves when both qualify: the one earlier in `W.Actors` runs first, shoves,
and stuns the other at once, which takes it out of the running. Consistent within
one simulation; slice order picks it.

### Only a fighter on the ground is shoved

A fighter in the air is skipped by `crowding` until it lands. Found by the bot
course test: a bot standing on a ledge shoved a climber out of the air every time
it jumped for the ledge, six times over, and then lost sight of it for good. A
shove is a ground move; one landing in another is shoved as any other overlap.

### Turn toward the crowder, then push it away

`enterPushState(other)` faces the crowder, then the shove moves it away from the
shover:

```
  before (A standing, facing right)     after (A facing left)

      B  A -->                            B          <-- A
    [==][==]                            [==]  . . . [==]
    dx in [-13, 13]                     |dx| in [16, 29]
                                        apart, face to face, both in reach
```

For an overlap in front the shover already faces the crowder. On the same column
(`dx = 0`) there is no side, and the shover's own facing decides.

The alternative - shove the way the fighter already faces, throwing a crowder
behind it through its body to its front - lands inside reach too, but `push.png`
shows both arms pushing forward, and the crowder passing through the body to meet
them reads as a teleport.

For a player, the facing the shove sets holds while it plays; on returning to
standing the mouse decides again.

### `PushDistance` = 16, bounded both ways

The crowder starts 0 to 13px from the shover and is moved `PushDistance` further:

```
  must clear the overlap:   0 + PD >= 14   ->  PD >= 14
  must stay in reach:      13 + PD <  35   ->  PD <= 21

   14 ...... 16 ............. 21
```

16 lands the crowder 16 to 29px out. Both bounds are asserted by a test sweeping
every overlapping offset, both facings, not by the comment beside the constant.

### A remembered crowder stops a shove repeating

A wall can cut the throw short, leaving the crowder where it was, still
overlapping; and the glide leaves it overlapping for its first tick or two in any
case. Nothing else stops the next tick on the shover's feet from throwing the
same shove again, every half second, with no other button read (found by
`TestAWallStopsTheThrow`). So the fighter remembers, in `PushStuck`, the fighter
its shove landed on, and `crowding()` skips that one until the two are apart. Usually the crowder, once its stun ends, shoves back and parts them
itself; the memory covers the one that does not, such as one guarding everything.

It is recorded on the landing frame in every world, not only the authority's. A
client moves nobody, so on a client the crowder is always still overlapping and
always remembered - which is what stops a predicting client throwing the same
shove every half second while it waits for the snapshot that parts them.

### A glide in whole pixels, not a jump and not a speed

The shove carries the crowder rather than putting it down 16px away in one tick,
which read as a teleport. It sets `ShoveLeft` and `ShoveDir` on the crowder, and
each tick after the crowder's own move `glide` takes a third of what is left,
rounded up: 16 goes 6, 4, 2, 2, 1, 1, an ease-out over six ticks that spans the
arms-out frames 3 and 4.

Speed was the obvious alternative and does not work here. Position is truncated
to whole pixels every tick, so anything under 60 px/s moves nobody - the bug
recorded in `flinch-knockback-is-lost-to-rounding` - and a slide under friction
loses its whole slow tail to it; its distance would be neither 16 nor anything
the bounds could be written against. The glide moves whole pixels only, so it
travels exactly `PushDistance`.

Each step asks `blockSides`, as every sideways move does; a step a wall stops
short ends the glide there, flush, with the rest dropped rather than pressed into
the wall. The crowder's own speed is zeroed when it is stunned, so a runner
stunned mid-stride does not slide on into the shover during the wind-up.

The glide belongs to the crowder, not to the shover's state: a shover hit out of
its shove after landing does not leave the crowder half-way.

### The crowder is stunned when the shove starts, not when it lands

Found in play: the shove rarely finished. A reproduction with a real bot
walking onto a standing player's back showed two ways it was cut off. Once it
landed, the crowder - now face to face and in reach - swung, the shover's guard
matched the height, and the provisional block ended the shove at frame 4 of 7.
And before it landed, the crowder was usually already mid-swing, because the
lunge of each swing is how a bot walks in; that swing arrived inside the 9-tick
wind-up and the shove never threw anyone.

So the shove stuns, and stuns at once. `enterPushState` puts the crowder into
`enterShovedState` on the authority: `StateAttacked` and `hit-top.png`, the
lockout of a stun, ending whatever it was doing. It is not `enterAttackedState`:
it skips `hitFeedback`, so there is no structure cost, no hit counted, no flash,
no `EvHit` and no hit-stop, and its speed is the shove's `PushSpeed` rather than
`StunKnockback`, which friction would carry about 45px - past the 35 the
distance bounds are there to keep. A fighter already stunned keeps the stun it
has rather than having it restarted.

Stunning only on the landing frame was tried first: it fixed the first failure
and not the second. After the change, 15 shoves across 20 randomised runs all
played to the end.

### Reach is asked in one place

`swingReachFacing(t, o)` is lifted out of `inSwingRange` with its block-all
fallback (measure the swing the fighter would throw the moment it stopped
guarding). `inSwingRange` becomes a one-line caller, and the shove's distance
tests measure against it.

### Carried over unchanged from the withdrawn version

`StatePush` and `push.png` appended; `PushFrame` 3, the frame the arms come out; the `PushDone` once-latch; the dead neither shove nor
are shoved; the displacement lands only where `W.Authority`; `StatePush` excluded
from `canMove`; `jump_test.go`'s runners spaced apart. Not carried over: the
one-tick displacement and `PushSpeed`, replaced by the glide.

### Ten frames at three ticks each

`push.png` grew from 7 frames to 10: three more recovery frames on the end, back
to standing, with the arms still first fully out on frame 3 (ink to x=45) and
the follow-through on frame 4 (x=47). The rate is kept at three ticks a frame,
so `PushSeconds` goes from 0.35 to 0.5 and the shove lands at the same moment,
0.15s in; only the recovery is longer.

The crowder's stun is 0.3s, so it recovers about 12 ticks before the shove's
last frame. In the reproduction, 23 of 25 shoves played all ten frames; the
other two lost their last recovery frames to the crowder's next swing meeting
the shover's guard, long after the throw had landed.

## Risks / Trade-offs

- **[A client predicts a turn the server does not make]** -> The crowding check
  is not authority-gated, only the displacement and the stun are, so a
  predicting client turns and plays the shove on its own reading of a lagged
  remote fighter. Measured in play under latency simulation (tasks); if it
  shows, the check moves behind the authority and the client plays whatever
  `StatePush` the snapshot carries, at the cost of a round trip.
- **[Close fighting turns into shoving]** -> A bot's lunges walk it into its
  target a few swings into any exchange, and the overlap is shoved - by the bot
  as often as by its target. That stuns one of them for 0.3s and resets the
  distance. It is what was asked for; if it makes fights feel stop-start, the
  lever is `AttackDash` or how the bot spaces itself, not the shove.
- **[A stun for every overlap]** -> Walking into anyone who is free costs a stun.
  In a chase that is the pursuer, whenever it catches up.
- **[Tests assumed a quiet world]** -> Anything that stands two fighters within
  14px of each other now has a shove in it. Swept by running the suite
  repeatedly, not by inspection.

## Migration Plan

`StatePush` and `push.png` are appended, so no existing state or animation id
moves. Client and server ship together: an old client decoding a snapshot with
`StatePush` would reject it as out of range. Rollback is reverting the change.
