## Context

See proposal.md — Why. The state that matters for the approach:

```
  latchInput()        game.go:303       gated !spectating && !matchHeld()
    |
    v
  pressedAttack / pressedJump           actor.go:810, package-level bools
    |
    | sampleInput() clears at actor.go:844
    v
  Input.Attack / Input.Jump
    |
    +-- offline --> game.go:245   actor.Input = sampleInput(actor)
    +-- client  --> net_client.go:310   a numbered, ringed, sent command
                    |
                    v
            Actor.Update  actor.go:228
                    |
            FreezeTicks > 0 ? --- yes --> actor.go:230 return   * INPUT DIES *
                    |
                    no
                    v
            state switch, six Input reads:
              288  StateIdle  Attack        291  StateIdle  Jump
              310  StateRun   Attack        307  StateRun   Jump (run-jump)
              1163 StateAttack Attack -> AttackQueued  (bool, no expiry, replicated)
              1353 land()     Attack
```

Three facts shape everything below, all verified against the tree rather than assumed:

- **`sampleInput` is not input-layer only on a client.** It has two callers; `net_client.go:310` mints the tick's networked command from it (`Seq`, `Ring`, `send`, and replay on reconcile). Anything done there is done to the wire.
- **`Flags uint16` has three free bits.** Bits 0–7 are the named flags (`net_protocol.go:129`), 8–9 are `HitCount` (`fHitCountShift = 8`), 10–12 are `HitRun` (`fHitRunShift = 10`). Bits 13, 14, 15 are unused, and `actorRecSize` is 44.
- **A fighter's energy cannot move during its own freeze.** `regenEnergy` is at `actor.go:243`, below the return at `:230`, and spending only happens through actions that are equally gated.

## Goals / Non-Goals

**Goals:**

- Fix attack and jump in one mechanism, since they share the hole.
- Put the buffer where prediction and rollback already handle state, so client and server cannot disagree by construction.
- Take no new tunable. The freeze already has a length; the buffer should inherit it rather than introduce a second number to balance against it.
- Leave the six `Input` read sites, `sampleInput`, and both paused modes untouched.

**Non-Goals:**

- See proposal.md — Non-goals. At design level, additionally: no change to the shape of the latch (still two package-level bools), no change to `AttackQueued`, and no growth of any packet.

## Decisions

### Buffer inside `Update`, folded back into `a.Input` before the switch

The freeze guard captures instead of dropping, and the first live tick restores:

```
  frozen tick (FreezeTicks > 0):        first tick after the freeze:

    FreezeTicks--                         Input.Attack |= BufAttack
    BufAttack |= Input.Attack             Input.Jump   |= BufJump
    BufJump   |= Input.Jump               BufAttack, BufJump = false
    return                                ... then the state switch, unchanged
```

Folding back into `a.Input` rather than teaching each consumer about the buffer is what keeps the blast radius at one function. Every state arm keeps its own gate — `a.Updown != 3` at `:288`, `enterRunJumpState()` at `:307`, `canSwing()` inside `enterAttackState` — and none of them can tell the press was late.

*Why not hoist the `AttackQueued` assignment above the freeze return* — the first shape this took, and it does not work. `AttackQueued` is only *read* under `case StateAttack` (`actor.go:1233`). A flinched fighter is in `StateIdle` or `StateRun`, which read `a.Input.Attack` directly, so the hoisted buffer would never be consumed there and would sit `true` until some later swing chained off it. That is a second bug, not a one-line fix.

*Why not a `frozen()` helper beside `matchHeld()`* — the earlier design added one to make the two predicates read as a matched pair. Under the routing/buffering split they are not a pair: `matchHeld()` answers "does another consumer own the device", `FreezeTicks` answers "can this fighter act". The guard at `:230` already reads plainly, and a helper whose only job is to imply a symmetry that does not exist is worse than none.

### The lifetime is the freeze, not a window

The buffer is set only on a frozen tick and cleared on the first unfrozen one, so its life is bounded by `HitStopTicks` — 6 ticks, 100 ms — with no constant of its own.

This is the whole reason to prefer it to a general input buffer. A decaying buffer needs a window; every candidate window is wrong for something. Six to eight ticks covers hit-stop but is far shorter than the 18 ticks `AttackQueued` effectively grants inside a 0.3 s first swing, so chaining would go from "click any time" to "click late", which is a feel change. Eighteen or more preserves chaining but gives every other state a long tail of stale presses. Two windows is honest but is no longer one model. Deriving the lifetime sidesteps the choice entirely, and leaves the general buffer available later as a deliberate design pass rather than a side effect of a bug fix.

### Replicate in two free flag bits

`fBufAttack` and `fBufJump` at bits 13 and 14, set in `recOf` beside `fAttackQueued` and read in `apply`.

Replication is not optional here. `reconcile` (`net_client.go:283`) calls `r.apply(local, c.W)` and then replays the ring from `ack+1`. If the buffer is not in the record, `apply` zeroes it, and a freeze that started before the acknowledged tick loses its buffered press on every correction — the bug back again, only intermittently and only under loss, which is the worst version of it. It also has to be in the record for `recOf(local) != before` at `:294` to compare like with like.

The cost is two bits in a field that already exists. `actorRecSize` stays 44, *Still fits the packet* is untouched, and no snapshot gets larger. `ProtocolVersion` still moves 6 → 7 because the bits change meaning: an old peer would read them as zero and silently lose predicted presses rather than fail loudly.

*Note for sequencing:* `fix-client-respawn-handoff` also needs a version bump, and its is a real header growth. If both land, bump once and decide the order first, or the second change to merge will be bumping a version the first already moved.

### The stun boundary needs no code

The fold-back happens on the first unfrozen tick whatever the state. For a stunned fighter that state is `StateAttacked`, whose arm calls `updateAttacked()` (`actor.go:316`) and reads no input at all — so the press is restored and dropped by that same tick, and cannot resurface when the stun animation ends. For a flinched fighter the state is `StateIdle` or `StateRun`, which do read it.

So the spec's split — a flinch does not lock you out, a stun does — is produced by the state machine rather than by a condition in the fix. Worth a test precisely because it is emergent: nothing in the buffer code says "except during a stun", and a future arm that starts reading `Input` would change the answer without touching this change.

### Death clears it

`enterDeathState` (`actor.go:1063`) already zeroes `AttackQueued`, `HitCount`, `HitRun`, `FlinchTicks` and `DeathTicks`; `BufAttack` and `BufJump` join that line. Without it, a press buffered on the killing blow outlives the fighter, and — given `fix-client-respawn-handoff` is about to make respawn actually hand control over — would fire on the new fighter's first tick.

### Testing without a window

Reachable without a GL context, the way `structure_test.go` already drives combat: build actors, set `FreezeTicks`, set `a.Input` directly, call `Update`, assert on `BufAttack`/`BufJump` and on the resulting state. The device read in `sampleInput` needs raylib, so the tests set `Input` rather than going through `latchInput`. Note that several existing tests zero `FreezeTicks` deliberately (`structure_test.go:32`, `:203`, `:218`, `:230`) to skip the freeze; they keep passing untouched, and the new tests are the ones that exercise it.

## Risks / Trade-offs

- **Two more bits of replicated per-actor state.** → They are inside an existing field, so nothing grows, and they are cleared within 6 ticks by construction. The exposure is the version bump, which the handshake already enforces.
- **The fold-back is invisible at the read sites.** A reader at `actor.go:288` sees `a.Input.Attack` and cannot tell it may be a press from up to six ticks ago. → The comment at the guard is the place to say so; it is one location rather than six, which is the trade being made.
- **The stun behaviour is emergent, not stated in code.** → Pinned by a scenario and a test, and called out above so the next person to add an input read to `StateAttacked` knows they are changing it.
- **Client and server must agree on the freeze for the fold-back to land on the same tick.** → `FreezeTicks` is already replicated (`net_protocol.go:181`, `:236`) and both sides run the identical guard, so this is exact rather than probabilistic — which is the main advantage over withholding the latch. Still worth an eyes-on check under `F5`–`F8` latency simulation rather than trusting it.
- **Pre-existing, not introduced here:** during the respawn gap there is no `IsPlayer` actor, so nothing samples and the package-level latch is never consumed — a click held from before death fires as the new fighter spawns. That is the latch, not the buffer, and this change neither causes nor fixes it; `fix-client-respawn-handoff` is where it belongs.

## Migration Plan

`ProtocolVersion` 6 → 7. No saved data, no file format, no packet-size change. Client and server must be built together; the existing handshake refuses a mismatch. Revert is the single condition — dropping the two fields and the two flag bits restores the old behaviour exactly.
