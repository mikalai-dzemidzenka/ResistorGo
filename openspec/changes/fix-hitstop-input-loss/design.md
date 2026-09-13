## Context

See proposal.md — Why. The state that matters for the approach:

The latch is one pair of package-level bools in `actor.go`, outside the simulation. Two places turn it into a tick's command, and both hand it to an actor that may be frozen:

```
  game.go:222     actor.Input = sampleInput(actor)     offline, per IsPlayer actor
  net_client.go:309   in := sampleInput(local)         client, once per tick

  actor.go:765    sampleInput: pressedAttack, pressedJump = false, false
  actor.go:198    Update:      if a.FreezeTicks > 0 { a.FreezeTicks--; return }
```

`sampleInput` already takes the actor it is sampling for, so the freeze is in scope at the only line that needs to change. Three constraints shape everything below:

- **The client must emit one command per tick.** `replication` — *Input commands* requires it, and `net_client.go:307` derives `Seq` from it. Nothing may skip sampling.
- **`FreezeTicks` is already replicated** (`net_protocol.go:170`, `:221`), so client and server reach the same verdict on a given tick without new protocol.
- **The server never samples.** `sampleInput` has exactly two callers, both client-side; the server replays received commands. The fix cannot desync by construction.

## Goals / Non-Goals

**Goals:**

- One change site, chosen so that offline and networked play get the fix from the same line.
- No packet-format, snapshot-flag or replication change.
- The two "is something holding this?" predicates stay legible as separate questions.

**Non-Goals:**

- See proposal.md — Non-goals for scope. At design level, additionally: no change to the shape of the latch (still two bools, still package-level, still global), and no new simulation state. Anything that needs replicating or rewinding is out.

## Decisions

### Withhold the latch, do not skip the sample

`sampleInput` clears the latch only when the sampled actor is not frozen. It still reads the devices, still returns a full `Input`, still produces a command.

*Why not skip `sampleInput` on a frozen tick* — it was the first shape of Option A and it is wrong: no command that tick breaks the one-per-tick rule and puts a hole in `Seq`.

*Why not gate in the two callers* — `stepOffline` and `Client.tick` would each carry their own copy of the condition. That is the pattern `fix-paused-input-latch` removed, and the reason time manipulation was missed the first time. One site, in the function that already has the actor.

### `Attack` reads true for the whole freeze, and that is fine

The latch stays set, so the flag repeats on each frozen tick and the first live one. Each consumer is safe under repetition: a frozen tick ignores `Input` entirely; `enterAttackState` fires at most once because the state leaves `Idle`/`Run`; `updateAttack` setting `AttackQueued = true` is idempotent. The spec covers this as *One press is one action*.

*Why not send the edge exactly once* — that would mean deciding client-side which tick the edge belongs to, i.e. duplicating the freeze rule in the input layer. Letting the flag repeat and having the simulation pick it up puts the decision in one place, the state machine, where it already lives.

### A per-actor `frozen()` beside `matchHeld()`, not folded into it

Add a small `func (a *Actor) frozen() bool` and use it at both `Update`'s guard and the new condition. It is three lines and buys the thing that matters here: the two predicates sit side by side and read as different questions — `matchHeld()` global and discarding, `a.frozen()` per-actor and deferring — so the next pause path added to one is not silently assumed to belong in the other. proposal.md — *Why hit-stop defers where a paused mode discards* is the rationale; this is where a reader will look for it.

*Alternative rejected:* extending `matchHeld()` to cover freezes. It is global, so it would gate the whole fixed-step loop on one fighter's freeze and stop the world for both. Wrong answer and wrong granularity.

### Testing without a window

Both cases are reachable without a GL context, the way `rewind_test.go` and `devtools_dev_test.go` already reach theirs: set `FreezeTicks`, set the latch directly, call `Update`, assert on the latch and the resulting state. The device read in `sampleInput` needs raylib, so the tests drive `Update` and the latch rather than going through `latchInput`.

## Risks / Trade-offs

- **A press survives longer than before, so a stale click could fire late.** → Bounded by the freeze: the first live tick either acts on it or drops it per its own state. `HitStopTicks` is 6, so the ceiling is 100 ms. The stun case is pinned by its own scenario.
- **`Attack` repeating across ticks makes an "edge-triggered" field look level-triggered for that span.** → The wire format is unchanged and every consumer is idempotent, but `replication` — *Input commands* calls attack and jump edge-triggered, so a future consumer could read the name and assume one tick. The comment at `actor.go:730` is the place to say otherwise.
- **Server and client must agree on the freeze for the repeat to be harmless.** → They do: `FreezeTicks` is replicated and the server drops input on a frozen tick by the same early return. Worth an eyes-on check under `F5`–`F8` latency simulation rather than trusting it.
- **Pre-existing, not introduced here:** during the respawn gap there is no `IsPlayer` actor, so nothing samples and the latch is never consumed — a click held from before death fires as the new fighter spawns. `reap()` removes before `Spawn`, so there is never a second sampler, only zero. This change does not alter that, and it should not be read as a regression from it.

## Migration Plan

None. No data, no protocol, no saved format. Revert is the single condition.
