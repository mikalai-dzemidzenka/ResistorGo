## Context

See proposal.md for why. The tick's horizontal move is, in `Actor.Update`:

```
  fromX := a.Pos.X
  a.Pos.X += a.Speed.X * dt
  a.Pos.X = float32(int(a.Pos.X))    <- the floor: < 60 px/s moves 0
  a.blockSides(fromX)
```

`updateHorSpeed` applies A/D only when `canMove`, which excludes
`StateAttack`, and turns the fighter with `face` whenever a key is held. The
attack case then clamps `Speed.X` to `+-AttackDash` - the same constant as
the 4 px dash, used here as a speed of 4 px/s, which the floor always reduced
to nothing. Friction (`DefaultFriction` 250 px/s^2) runs whenever no key is
held; a walk's last 60 px/s of deceleration has always been invisible because
of the floor, and a flinch's 50 px/s never showed at all.

Drawing already rounds: `drawPos` rounds `InterpolatedPos`, which the
interpolation spec requires. Snapshots and the rewind carry `Pos` as a float.

## Goals / Non-Goals

**Goals:**
- Slow horizontal speeds move, everywhere, by dropping one line.
- Walking the same everywhere, stopping close to where it did on average, and
  knockbacks travelling the distance their speeds describe.
- A swinging fighter's movement is capped, keeps its facing, and needs no new
  state.

**Non-Goals:**
- Vertical sub-pixel position: vertical speeds are far above the floor.
- Moving while blocking, stunned, shoved or clashing.
- Bot tactics that use the creep on purpose (stepping back to dodge).
- Changing the attack dash.

## Decisions

### Drop the horizontal truncation

The truncation line goes. `Pos.X` keeps its fraction; everything that places a
fighter exactly (wall stops in `blockSides`, climb and ledge snaps, spawn
columns, the clash's placement) still writes the positions it writes. Nothing
reads `Pos.X` expecting an integer except tests that assert exact positions,
which are updated where a slow tail now shows.

Alternative considered: step one whole pixel every N ticks while swinging.
Keeps integers, but only for this one motion; the flinch stays broken and the
next slow speed hits the same floor with no warning.

### Stop a released walk at the old floor

```
  in updateHorSpeed, after friction, when no key is held:
    if |Speed.X| < WalkStopSpeed (60)
       and the speed is the fighter's own
         (State is Idle or Run, FlinchTicks == 0, ShoveLeft == 0)
    then Speed.X = 0
```

Without it every walk slides a further v^2/2f = 60^2 / 500 = 7 px after the
key is released. Only a fighter
moving under its own power is stopped: a flinch holds `FlinchTicks`, a stun is
`StateAttacked`, a shove holds `ShoveLeft`, so their speeds run out under
friction and now travel their full distance. `WalkStopSpeed` is 60 because
that is where the floor stopped walks before.

It cannot make a walk end where it did, because there was no single place a
walk ended. Cutting toward zero is not rounding down: for a fighter at
positive x, moving right lost the fraction every tick and moving left gained a
whole pixel, and past the origin the other way round. Measured before this
change, from a stand:

```
  1 s walk                 toward -x      toward +x
    at x = 100             134 px          78 px
    at x = -300             95 px         136 px
  slide after release      43 px          26 / 60 px

  without the cut: 123.6 px toward -x, 106.0 px toward +x on either side
  (the remaining difference is walking toward or away from the facing, an
  existing rule); slides 58.5 / 41.1 px, about 7 px less with the stop
```

The walk's pace is therefore a fix and a change at once, and is judged by eye
in play rather than pinned to the old numbers.

### Move during an attack, capped, without turning

`canMove` admits `StateAttack`. While attacking, a held key accelerates the
fighter as it would walking but does not call `face`, and the attack case's
clamp becomes `+-AttackMoveSpeed` (20 px/s) instead of `+-AttackDash`. The
chain's `Speed.X = Orientation * AttackDash` (4 px/s of forward drift at each
blow) stays within the cap and is left alone.

The player's mouse-facing rule in `updateHorSpeed` already turns an attacker
only on frame 0 of a swing; that is unchanged, so a player can still aim a
swing's first frame, and holding a key never turns anyone mid-swing.

Bots steer with the same `Input.Left`/`Input.Right`, so they creep when their
steering asks to - mostly while out of reach, since a bot at arm's length
holds still. The line and the clash place fighters by position and are not
affected; the clash already zeroes speed.

### Constants

```
  AttackMoveSpeed = 20   // px/s while swinging; ~6 px over a low swing
  WalkStopSpeed   = 60   // the old floor: a released walk stops here
```

`AttackDash` keeps only its meaning as the dash's distance.

## Risks / Trade-offs

- [Knockbacks lengthen] The stun's throw was measured at 20 px with its tail
  hidden. Measured after the change it ran 33 px, not the 7 px more guessed:
  the cut lost part of a pixel on every tick of the throw, not only in its
  tail. `StunKnockback` is lowered from 150 to 107, which throws 20 px again;
  the flinch now moves 4.6 px, about a quarter of that. The shove's glide is
  pixel-stepped and unaffected, but any friction after it now shows. →
  Re-measure both in a test and decide by eye whether `StunKnockback` needs
  lowering to keep the throw at its intended length.
- [Provisional block pins a dodger] A fighter guarding the height of an
  incoming swing is put into a block on the swing's first tick and cannot
  move there, so it cannot step out of that swing. That is the existing rule
  for blocking and is left as it is: distance decides a swing the fighter is
  not guarding.
- [Tests asserting whole-pixel positions] About 30 test lines compare exact
  `Pos.X`; those that follow friction tails may now see fractions. → Update
  the expectation, or compare within a pixel, only where the new value is the
  correct one.
- [Prediction corrections] A client and server now agree to a fraction
  instead of a whole pixel; reconciliation already compares floats, and the
  correction is smoothed, so the only effect expected is corrections becoming
  rarer, not larger.
