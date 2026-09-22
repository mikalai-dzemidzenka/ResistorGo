## Context

See proposal.md for why. A swing starts in `enterAttackState`, called from
`StateIdle`, `StateRun` and `land()`; chained blows are set up inside
`updateAttack` without it, which is what makes "an opening swing"
identifiable. `updateAttack` resolves hits on the authority: an opening low
swing whose box overlaps a fighter guarding low puts that fighter into a
provisional `StateBlock` on the swing's first update, whatever that fighter
was doing, including an attack of its own. `updateBlock` then settles it at
the hit frame. The low swing's hit frame is past frame 5 of 8 over 0.3 s,
about 13 ticks in.

```
  tick 0        A opens low swing            (enterAttackState)
  tick 1        A.updateAttack: B guarding low and in the box
                -> B forced into StateBlock (provisional)
  tick 1..6     B presses attack: ignored today (updateBlock reads no input)
  tick ~13      hit frame: B's block confirmed
```

So a second swing a few ticks late never becomes a swing at all. The clash
has to be caught in two places: A's `updateAttack` finding B mid opening
swing, and B's `updateBlock` finding its own attack press against A.

## Goals / Non-Goals

**Goals:**
- One function decides and starts a clash, whichever of the two places finds
  it, so the conditions cannot drift apart.
- Hit resolution, the shove and the state machine each learn about the clash
  in one line, the way they already know about `StateAttacked`.

**Non-Goals:**
- Mid and high clashes; their sheets do not exist yet. The trigger is written
  so that allowing another height is a change to one condition and an
  animation lookup.
- A sound or camera effect for the clash.
- Predicting a clash on the client.

## Decisions

### Record the tick a swing opened

`Actor` gains `SwingTick uint32`, set to `a.W.Tick` in `enterAttackState`
(nil world in tests: left alone). An opening low swing is `State ==
StateAttack && Anim == attack-low.png`: the chain replaces the animation with
`attack-low-2.png` or an `updown-atk-*` sheet, so the check needs no extra
flag. `SwingTick` is not replicated: only the authority reads it. Being an
ordinary value field, time manipulation records it with the rest.

Alternative considered: a countdown field decremented each tick. A tick stamp
cannot drift with hit-stop freezes, which pause one fighter's update and not
the world's tick, and a 100 ms window is a statement about the world's time.

### `tryClash(a, b)` is the single gate

```
  tryClash(a, b) bool     authority only
    - not allies, both on the ground, neither Clash/Attacked/Death
      (a shove's stun is StateAttacked too)
    - each faces the other: facing(b.Pos.X - a.Pos.X) == a.Orientation, and
      the same from b
    - each in the other's reach: a.inSwingRange(b) && b.inSwingRange(a)
    - a is in its opening low swing, and b is either
        * in its opening low swing, |a.SwingTick - b.SwingTick| <= ClashWindowTicks
        * in StateBlock with Attacker == a, pressing attack at Updown 0, able
          to pay, and W.Tick - a.SwingTick <= ClashWindowTicks
    -> startClash(a, b)
```

Callers:
- `updateAttack`, at the top of the authority branch, before the hit/block
  loop: for each other fighter, `tryClash(a, other)`; on success, return.
  Because this runs before the loop that would throw the other into a block,
  a same-tick pair never reaches the block at all.
- `updateBlock`, first thing: if `a.Input.Attack` and `a.Attacker != nil`,
  `tryClash(a.Attacker, a)`; on success, return.

The arguments are ordered so `a` is always the fighter in its opening swing;
`startClash` itself orders the pair by position.

### `startClash` pays, places and starts both

1. Pay: `b` pays `EnergySwing` if it came from a block (its swing was never
   paid); both `spendStructure(ClashStructure)`. A fighter that breaks enters
   death with the other as the killer (`enterDeathState(other)` already
   refills the killer's energy); if either broke, the survivor goes to idle
   and nothing else happens.
2. Place: left = the one with the smaller `Pos.X`. Both at the midpoint
   `ClashGap / 2` either side, each through `blockSides` from where it stood
   so a wall stops it, left facing right and right facing left, `Speed`
   zeroed.
3. Start: both `State = StateClash`, `stopFlicker()`, `AttackQueued`,
   `BufAttack`, `BufJump`, `SwingHit` cleared, `Attacker` set to the other,
   and `SetAnim` with `reset: true` on the same tick: left
   `battle-low-1-left.png` over 9 frames, right `battle-low-1-right.png` over
   10, both over `ClashSeconds`.

`ClashGap` is where the two sheets' blows line up. It starts at the low
swing's reach and is tuned by eye in the game. `ClashSeconds` starts at 0.7 s
(the sheets hold 9 and 10 frames against the swing's 8 in 0.3 s) and is
tuned the same way.

### The state

`StateClash` is appended to the `ActorState` enum before `stateCount`, and
its case in `Actor.Update` zeroes `Speed.X`, ignores input and goes to
`enterIdleState` when `AnimFinished`. The shove gate already admits only
`StateIdle`/`StateRun`, so a clashing fighter never shoves; `crowding` also
skips a clashing `other`, so it is never shoved. `updateAttack`'s target loop
skips `StateClash` alongside `StateAttacked`, which covers hit, block and
provisional block. `face` refuses a turn in `StateClash`, as in `StateClimb`.

The two sheets are appended to `animTable` and registered in `LoadTextures`.
Neither has an entry in `attackFrames`: a clash is not an attack, and
`updateBlock`'s "attacker still swinging" logic never sees one because a
clash ends any block in it.

### Constants

```
  ClashWindowTicks = 6     // 100 ms, well inside the ~13 ticks to the hit frame
  ClashStructure   = 0.5   // half a hit
  ClashSeconds     = 0.7   // tuned in game
  ClashGap         = ...   // tuned in game, starts at the low swing's reach
```

The energy is `EnergySwing` itself, not a separate constant: the clash costs
the swing that opened it.

## Risks / Trade-offs

- [Update order] A fighter updated after its partner sees the partner already
  in `StateClash` on the tick it starts, and skips it. Both are placed and
  started by the one `startClash` call, so neither fighter's own update has to
  do anything that tick. The other one's update that tick runs the clash case
  and only zeroes speed.
- [Client shows a swing first] A predicting client plays its own fighter's
  swing until the snapshot arrives with the clash; reconcile resets it and the
  replayed commands do nothing in `StateClash`. The clash appears one round
  trip late, with a short swing before it. Accepted, as the shove already is.
- [The sheets may not line up at any single gap] Their art decides the gap.
  The first in-game task checks it before the numbers are settled.
- [A window that feels too tight or too loose] It is one constant; the
  in-game task sets it.
- [Hit-stop] A fighter frozen by hit-stop (`FreezeTicks`) does not update,
  so it cannot open a swing while frozen. That is the existing rule for any
  action, and the window is measured in world ticks so a freeze on one side
  does not widen it.
