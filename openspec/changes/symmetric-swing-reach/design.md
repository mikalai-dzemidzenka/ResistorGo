## Context

See proposal.md - Why.

Three rectangles matter, all expressed relative to a fighter's `Pos.X` (call it
`P`). Measured from the sprite sheets and from `NewActor`:

```
  48px sprite frame:   [P ........................... P+48]   centre P+24
  drawn body (idle):        [P+10 ............ P+35]          centre P+23
  collision box:                [P+14 ... P+28]               centre P+21
  attack box (right):                [P+24 ......... P+48]
  attack box (left):   [P ......... P+24]
```

`attackBoxFacing` builds the attack box as `pos.X + box.X + 12*(o+1)` with
`box.X = 0` and `box.Width = 24` for all three guard heights - exactly the front
half of the 48px cell, mirrored about the frame's centre at 24. The art agrees:
`attack-mid` frames 5-7 have ink out to x=47 and `push` frame 3 to x=45, so the
arm genuinely reaches the frame edge, and mirroring puts it at x=0 facing left.
**The attack box is correct.**

`CollisionBox` is `{X:14, W:14}`, centred at `P+21` - 3px left of the frame
centre. Every reach test (`swingReachFacing`, and `updateAttack` via
`attackBoxAt`) compares the frame-mirrored attack box against this
frame-off-centre collision box, so the 3px offset enters with one sign facing
right and the other facing left.

With `dx = other.Pos.X - self.Pos.X` and `AttackDash = 4` applied:

| | near limit | far limit |
|---|---|---|
| overlap | `dx > -14` | `dx < 14` |
| reach facing right | `dx > 0` | `dx < 38` |
| reach facing left | `dx < 6` | `dx > -32` |
| crowding (overlap and not reach) | `(-14, 0]` | `[6, 14)` |

`dx = 0` falls inside reach facing left and outside it facing right, because
`aabb` is exclusive and the right-facing box starts at `P+28`, exactly the
collision box's right edge.

The shove turns that into a lock. `blockSides` clamps a shoved fighter to
`Pos.X = right - w - l`, which for a shover standing flush against a wall at `W`
(so `A = W-28`) evaluates to exactly `A`. The crowder lands at `dx = 0`, which
facing right is still crowding, so the shove is thrown again from an identical
position, forever. The crowding check in `Actor.Update` runs *before* the state
switch, and `case StatePush` reads no input, so the `StateIdle` branch that reads
`Input.Jump` and `Input.Attack` never executes. No input escapes.

Flush is not a rare coordinate: `blockSides` is the code that produces it, and
positions are truncated to int each step, so it is the normal resting state of
any fighter that walked into a wall.

## Goals / Non-Goals

**Goals:**

- Make `attackBoxFacing` mirror about the collision box's centre, so reach is a
  single number rather than one per facing.
- Keep `attackBoxFacing` the sole place reach geometry is expressed, so hits,
  the outline, lag compensation, bot approach and the shove cannot drift apart.
- Leave `crowding`, `inSwingRange` and `updateAttack` unchanged in structure -
  they consume the corrected box and need no special cases.

**Non-Goals:**

- Redesigning the shove. Options considered during exploration - shoving
  yourself back when the throw is clamped, or splitting the displacement between
  both fighters - are separate changes.
- Moving the collision box. `CollisionBox.X` is consumed by `checkCollision`,
  `blockSides`, `headBlocked`, ledge climbing and spawn spacing; re-registering
  it would shift every fighter 3px against the level geometry.
- Fixing the 1px art offset. See proposal.md - Impact.

## Decisions

### Mirror the attack box about the collision box's centre

Introduce `C = CollisionBox.X + CollisionBox.Width/2` (21) and a single
parameter `d`, how far in front of `C` the swing begins:

```
  facing right:  X = P + C + d
  facing left:   X = P + C - d - AttackBox.Width
```

The current code is this expression with `d = +3` facing right and `d = -3`
facing left. Making `d` one value is the whole fix.

Alternative rejected: special-casing `dx = 0` inside `crowding`. That would stop
the deadlock without correcting the reach, leaving hits, the outline and lag
compensation still disagreeing by 6px between facings, and leaving the next
consumer of `swingReachFacing` to rediscover it.

### `d = 0`

| | `d = +3` (keep right) | `d = 0` (split) | `d = -3` (keep left) |
|---|---|---|---|
| attack box right | `[P+24, P+48]` (unchanged) | `[P+21, P+45]` | `[P+18, P+42]` |
| attack box left | `[P-6, P+18]` | `[P-3, P+21]` | `[P, P+24]` (unchanged) |
| max reach, both ways | 38 | 35 | 32 |
| reaches `dx = 0`? | no, in neither facing | yes, in both | yes, in both |
| crowding window | `(-14, 0]` / `[0, 14)` | `(-14, -3]` / `[3, 14)` | `(-14, -6]` / `[6, 14)` |

`d = +3` is the only option that changes nothing about how the game currently
plays facing right, and it is the wrong one: holding the longer reach fixed means
holding the frame-mirrored box fixed, which keeps `dx = 0` out of reach in *both*
facings and extends the deadlock to left-hand walls as well.

`d = 0` is chosen over `d = -3` because the two current values are wrong by the
same 3px in opposite directions; averaging them is the only choice that does not
privilege an arbitrary facing, and it costs 3px of forward reach rather than 6px.
It leaves an 11px crowding window on each side, clear of the `dx = 0` that the
wall clamp manufactures.

In code this is `box.X = -3` rather than `0` in the three `AttackBox*`
rectangles, or equivalently `12*(o+1) - 3` in `attackBoxFacing`. The second is
preferred: it keeps the correction next to the mirroring term it corrects, where
a reader can see why it is there, rather than distributing it over three
unrelated constants.

### Re-derive bot approach distances rather than preserving them

`inSwingRange` now returns a different answer, and `tooClose` sits beside it with
its own hand-tuned constants. Rather than compensating to keep the bot's stopping
distance numerically identical, treat the new reach as the truth and re-check the
bot against it. A bot tuned to stop at the right distance facing right will
otherwise stop 3px late.

## Risks / Trade-offs

- **[The forward swing is 3px shorter and players notice]** -> The change is
  small relative to `AttackDash` (4) and to the 14px collision box, and it is
  compensated by 3px more reach in the other direction. Verified by playing, not
  by test - reach is a feel number.
- **[Client and server disagree during a rolling deploy]** -> The reach constant
  is compiled into both. A client on the old constant predicts a hit the server
  denies, or vice versa. There is no wire-format change, so this does not break
  connections, but both sides must ship together.
- **[Bot regression the course test does not catch]** ->
  `TestABotFollowsAFighterTheLengthOfTheCourse` holds the bot's swing back, so it
  exercises pathing rather than approach. A bot that stops at the wrong distance
  passes it. Covered by adding an approach-distance assertion rather than relying
  on the course.
- **[The deadlock is only reasoned about, never reproduced]** -> Every existing
  push test runs on `floor()`, a single row of tiles with no wall anywhere, so
  `blockSides` can never clamp. A regression test needs a level with a vertical
  face, which none of the current helpers build.
- **[`d = 0` is right for the current numbers, not structurally]** -> It works
  because 3 < 14, so `dx = 0` lands inside the symmetric reach window. A future
  change to `AttackDash`, `AttackBox.Width` or `CollisionBox.Width` could close
  that margin again. The shove-repeat guard in tasks is the structural answer;
  this change is the correct one.

## Migration Plan

No data, wire format or save format changes - the packet layout is untouched and
`attackBoxFacing` is derived, never serialised. Ship client and server together.
Rollback is reverting the constant.

## Open Questions

- Whether `tooClose` needs its own adjustment or whether it falls out of the
  corrected reach. Answerable while implementing; it changes neither the specs
  nor the approach.
