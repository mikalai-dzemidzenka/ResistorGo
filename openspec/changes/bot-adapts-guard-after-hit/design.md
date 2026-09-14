## Context

See proposal.md — Why. The state that matters for the approach:

```
  who sets Updown            actor.go:882  updateUpdown()
    a.Updown = a.Input.Updown          -- every tick, from the command
    a.PrevUpdown = old                 -- only when it actually changed

  where the command comes from
    human   sampleInput   actor.go:846  Updown derived from the cursor's height
    bot     botInput      actor.go:791  in := Input{Updown: a.Updown}   <- carries forward
                          actor.go:800  1 in 61 ticks: in.Updown = rand.Int() % 3

  where a hit is resolved
    updateAttack  actor.go:1174   if a.W.Authority && !a.SwingHit { ... }
      -> takeHit  actor.go:1004   death | stun | flinch
```

Three facts decide the design, all checked against the tree:

- **`botInput` carries the guard forward, `sampleInput` recomputes it.** A bot's next command repeats `a.Updown`, so a value written straight onto the actor persists. A human's next command is derived from the cursor, so the same write is gone on the next tick.
- **`takeHit` runs on the authority only.** A client never resolves a hit against itself, so a roll placed there cannot be predicted, cannot diverge, and cannot need reconciling.
- **`Updown` and `PrevUpdown` are already replicated** (`net_protocol.go:181`, `:240`). A guard change reaches every client through the ordinary snapshot.

## Goals / Non-Goals

**Goals:**

- One reaction, in one place, with no new state on `Actor` and nothing new on the wire.
- Bot-only by construction *and* by an explicit test, not by relying on a human's cursor to undo it.
- Leave the existing guard wander exactly as it is, so the two can be told apart in play.

**Non-Goals:**

- See proposal.md — Non-goals. At design level: no bot state machine, no memory of past hits, no per-bot difficulty, and no change to `confirmBlock`.

## Decisions

### The roll lives in `takeHit`, past the death return

```
  takeHit(attacker)
    attacker.SwingHit = true
    cost, HitRun++
    if spendStructure(cost) -> enterDeathState; return     <- the killing hit leaves here
    *  the roll goes here  *                                <- so a stun and a flinch both reach it
    HitCount++
    if HitCount >= StunEveryHits -> enterAttackedState; return
    enterFlinchState
```

Placing it between the death return and the stun count is what makes "a flinch and a stun adapt, a death does not" fall out of the existing control flow rather than needing a condition of its own.

*Why not in `enterFlinchState`* — it would miss the stun, and the spec wants both.

*Why not in `hitFeedback`* — it is called by the flinch and the stun, which is right, but also sits below `enterDeathState`'s own concerns and is described as "what every landed hit does, stunning or not". Adding one fighter type's reaction to it would put bot behaviour inside the shared feedback path.

### Gate on `!a.IsPlayer` rather than letting the cursor undo it

An ungated roll is *almost* harmless — `updateUpdown` rewrites a human's guard from the cursor on their next command. Almost, because the write lands mid-tick on the server, and any other attacker whose hit is resolved later in the same actor loop reads `guarding := a.Updown == actor.Updown` (`actor.go:1187`) against the flipped value. In a two-fighter match that is unreachable; with three or more it is a rare hit judged against a guard the player never chose. The gate costs one condition and removes the class.

`IsPlayer` is the right flag: bots are `Spawn(pos, false)` on the server (`net_server.go:287`) and offline alike, players `Spawn(pos, true)` (`net_server.go:242`), and `game.go:244` already routes input on exactly this distinction.

### Keep `PrevUpdown` honest

`updateUpdown` maintains the invariant that `PrevUpdown` is the guard before the last change, and `case StateIdle` uses it to pick the `updown-<from>-<to>.png` transition. Writing `Updown` directly without touching `PrevUpdown` would leave the pair inconsistent and the transition animation drawn from a stale height. So the write follows the same shape as `updateUpdown`: set `PrevUpdown` only when the value actually changes.

### Odds as a constant, not a literal

`GuardAdaptOdds = 3`, read as one in three, beside `StunEveryHits` and the other combat tunables. It is a feel number that will be turned in play, and `rand.Int()%3` buried in `takeHit` would be the kind of literal nobody finds again.

*On `rand`* — `botInput` already uses the package-level `math/rand` for the wander and the swing odds, so this adds no new source of nondeterminism and no new import. Rewind is unaffected: `rewind.go` restores recorded state rather than re-simulating, so a roll is never replayed.

### Testing

`takeHit` is reachable headlessly — `structure_test.go`'s `swing()` already builds an authoritative world with an attacker mid-swing and a defender in the box, and `hit_run_test.go` drives `takeHit` directly. The odds are not seeded, so the tests assert on what the spec actually promises rather than on a particular draw: over many hits a bot's guard both does and does not move, a player's never moves at all, and a killing hit moves nothing. A test that pinned an exact sequence would be pinning `math/rand`'s stream, not the behaviour.

## Risks / Trade-offs

- **One in three may be too strong or too weak.** → It is a constant, judged in play, and the change is four lines to retune or remove.
- **A bot can adapt during a stun it cannot act out of.** → Deliberate, and specified: the guard it wakes up with is the one the hit taught it. Worth watching in play for whether it reads as unfair.
- **Two random sources now move the same value.** The wander and the adaptation can fight, and a wander immediately after an adaptation will look like the adaptation failing. → Accepted: the wander is one tick in sixty-one, so it is rare next to a hit, and keeping them separate is what lets either be tuned without the other.
- **Nondeterminism in the authoritative simulation.** → Already the case via `botInput`, and confined to the authority, so no client can disagree about it.

## Migration Plan

None. No data, no protocol, no saved format, no version. Revert is deleting the constant and the block.
