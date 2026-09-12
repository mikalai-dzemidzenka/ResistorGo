## Context

See proposal.md (Why) for motivation, and `specs/multiplayer/*` for the required behavior. The current code has several properties that block networking. Each one shapes a decision below.

- **Rendering drives simulation.** `DrawAnimOnce()` computes `AnimCurrentFrame` and its return value sets `AnimFinished` (`Actor.Draw`). Many state transitions in `Actor.Update` wait for `AnimFinished`, and hit timing reads `AnimCurrentFrame` (`attackFrames`). A headless server, or replaying commands for reconciliation, would therefore never finish an animation.
- **Simulation reads local-only globals.**
  - Guard height (`updateUpdown`) comes from `Cur.AtkPos` projected through `Cam`.
  - Facing (`MouseOnLeft`) comes from `MouseWorldPos`.
  - Bots aim at `Pl`.
  - Non-player guard changes use `math/rand`.
- **Simulation writes presentation globals.**
  - Attacks bump `Cam.ZoomSpeed`.
  - `StateAttacked` shakes `Cam.Offset` for any attacked actor.
  - `enterAttackedState` sets the global `PauseTime` hit-stop, which freezes the whole world.
- **Animation time is global.** `SetAnim` stamps `AnimPhysTimeStart = PhysicsTime`.
- **One global world.** `Actors` is package-level, and `updateAttack` iterates it. The level collision data (`LE.objectsIndexed`) is also global, but it is static during a match.
- **go-enet** exposes `Host.Service`, `Peer.SendBytes(data, channel, flags)` and RTT getters. It is cgo and not goroutine-safe, so all ENet calls stay on the main goroutine. Packet data must be copied before `packet.Destroy()`, as in `examples/example9`.

## Goals / Non-Goals

**Goals:**
- The existing `Actor` state machine runs unchanged in meaning on the server, in client prediction and offline.
- One server implementation for listen and dedicated modes.
- Single-player behavior stays identical. It becomes "one authoritative world, no network".

**Non-Goals:**
- Bit-exact determinism across CPUs or Go versions. Reconciliation absorbs drift.
- Predicting other players, and client-side hit detection.
- Delta compression, interest management, and more than about 16 fighters.

## Decisions

### D1. Simulation is made pure: animation in the fixed step, input as data, effects as events
- **Animation.** `Actor` gets a local tick counter `Ticks` (incremented once per `Update`). `SetAnim` records `AnimStartTick = a.Ticks` instead of `PhysicsTime`. At the end of `Update`, a new `a.advanceAnim()` computes `AnimCurrentFrame` and `AnimFinished` from `(Ticks - AnimStartTick) * FixedTimeStep`, using the logic now in `DrawAnimOnce`/`DrawAnim`. Draw functions become read-only.
- **Input.** `Input` becomes the network command: `Seq`, buttons (`Left`, `Right`, `Attack`, `Jump`), `Updown uint8`, `FaceLeft bool`, `ViewTick uint32` plus a `ViewFrac uint8`.
  - The client samples these in `sampleInput()`. It computes the cursor→guard-height mapping and the mouse side there, against the predicted actor position.
  - `updateUpdown`/`MouseOnLeft` read `a.Input` only.
  - Bots produce the same struct in `botInput(world, actor)`, including their random guard changes (server only).
- **Effects.** `Update` appends to `a.Events` (bitmask: `EvAttackStart`, `EvHit`, `EvBlocked`). The client's presentation code consumes events for the locally viewed fighter: zoom on attack, shake on hit. For remote fighters it consumes them from observed state transitions. Events emitted while replaying are cleared, not consumed.
- **Hit-stop.** It becomes per-actor `FreezeTicks`, set on attacker and defender. `Update` returns early while it is non-zero. With exactly two actors offline, this equals today's global `PauseTime` freeze. Online it avoids freezing uninvolved players. `PauseTime` is removed.
- **Alternative considered:** keep `Draw` driving animation and send the finished flag from the server. Rejected, because prediction and replay would still diverge on the client.

### D2. A `World` value holds per-simulation state; level stays global
```go
type World struct {
    Actors    []*Actor
    Tick      uint32
    Authority bool          // resolves hits; false for the client's predicted world
    History   *LagHistory   // server only
}
var SV, CL *World           // server world, client world (either may be nil)
```
`Actor` gets `W *World`, and `updateAttack` iterates `a.W.Actors`. Offline mode uses one `World` with `Authority=true` (the current `Actors`/`Pl` globals become `CL.Actors`/`CL.Local`). A listen server needs the server world and the client's view in one process, so a single global `Actors` cannot work. This is the smallest step away from the "globals, no DI" convention: `SV`/`CL` are still package globals. `LE` (level geometry) stays global and read-only while online.
- **Alternative considered:** the listen server spawns a child `-dedicated` process. That avoids `World`, but adds orphan-process handling and is harder to debug. Rejected.

### D3. Source-style per-player command execution (no clock alignment)
Like Source's usercmds, a human player's fighter advances only when one of that player's commands is executed. The server keeps a per-player queue ordered by `Seq` and drops duplicates or older commands. Each server tick it executes queued commands subject to a token bucket: +1 token per tick, maximum 8. That gives jitter tolerance and a speed-hack limit.
- During the server tick, the actors of other players are not advanced by a player's command. Only bots are advanced by the server tick itself.
- Being hit only changes the target's state. The knockback movement is applied when the target's own ticks run, which is Source's behavior.

Because the client advances the same actor by the same commands, reconciliation is keyed by `Seq`, and the client never needs to predict the server tick. The server tick is used only for snapshots, interpolation and lag-compensation history.
- **Alternative considered:** a lockstep-style design where the client runs ahead of the server tick (Overwatch-style), and the server repeats the last input when one is missing. Rejected: it needs clock-sync, time dilation and input-repeat mispredictions.

### D4. Transport and protocol
- **Setup.** ENet host with 2 channels: 0 is unreliable (commands, snapshots) and 1 is reliable (control). Default port 27015. `enet.Host.Service(0)` is drained each frame on the main goroutine.
- **Encoding.** Hand-written little-endian binary with a 1-byte message type, in the style of `example9/shared/protocol.go`, in `net_protocol.go`. Decoders check lengths and return errors. Callers drop bad packets.
- **Messages:**
  - `Hello{proto u16, levelHash u64}` (FNV-1a of `level_1.json` bytes)
  - `Welcome{playerID u16, actorID u16, serverTick u32}`
  - `Reject{reason u8}`
  - `Cmds{count u8, cmds...}`: the newest command plus the 2 previous ones
  - `Snapshot{tick u32, ackSeq u32, count u8, actors...}`
  - `Bye{reason u8}`
- **Actor record.** Full state, not delta, about 45 bytes:
  - identity and position: `id u16`, `x,y i16` (positions are already integers), `speed x,y f32`
  - state and stance: `state u8`, `flags u8` (orientation, onGround, animOnce, animReverse, animWhite, attackQueued, canClimbL/R), `updown`, `prevUpdown u8`
  - animation: `anim u8` (index into an anim table), `animStartFrame`, `animTotalFrames u8`, `animYOffset i8`, `animDuration` as `u16` ms, `ticks u32`, `animStartTick u32`
  - combat: `attackCount u8`, `freezeTicks u8`, `attackerID u16`

  16 actors take about 750 bytes, which is under the 1200-byte limit. The full record is sent for every actor because the local fighter needs complete state for reconciliation, and one record type keeps things simple.
- **Anim table.** Animation paths are sent as a `u8` index. The table is an explicit `var animTable = []string{...}` shared by client and server. `LoadTextures` iterates it, so a headless server needs no textures and the order is stable.
- **Alternative considered:** send the anim string. Rejected, as it is too large per actor.

### D5. Prediction and reconciliation
The client keeps a ring of 128 `{Input}` entries keyed by `Seq`. Each client fixed tick:
1. Sample input.
2. Store it and send `Cmds`.
3. Run `Local.Update` in `CL` with `Authority=false` (the attack test is skipped, so predicted attacks only animate).

On a snapshot:
1. Copy the server record into the local actor, including `Ticks`/`AnimStartTick`.
2. Drop history entries with `Seq ≤ ackSeq`.
3. Replay the remaining entries with events suppressed.
4. Compare the pre- and post-replay drawn position. If the delta is ≤ 8 px, add it to a decaying `CorrectionOffset` (about 100 ms). Otherwise snap. Count a correction when any field differs.

Remote actors in `CL` are updated from interpolation (D6), not simulated. They take part in the local actor's ledge and ground collision only through level geometry, as today (actors do not collide with each other).

### D6. Interpolation buffer
- **Buffer.** The client keeps the last 32 snapshots.
- **Render time.** `renderTick = estServerTick - interpDelay/dt`. `estServerTick` advances at wall-clock rate from the newest snapshot tick. It is nudged towards `latestTick + sinceArrival/dt` by at most ±5 % per frame and never moves backwards. `interpDelay` is 100 ms, clamped to at least 2 × 50 ms.
- **Interpolation.** Per remote actor, lerp `x,y` between the bracketing records and round. Discrete fields come from the older record. `Ticks` is lerped too, and the anim frame is computed from it with the same function as D1, so all frames show. If there is no newer record, extrapolate `pos += speed*t` for t ≤ 250 ms.
- **Jumps and spawns.** A record-to-record position jump greater than `MaxSpeed*gap + 16` px is treated as a teleport. Actor appear and remove events are resolved against `renderTick`.
- **Lag-comp input.** The current `renderTick` (integer plus fraction) is written into each command as `ViewTick`/`ViewFrac`.

### D7. Lag compensation
`LagHistory` is a 64-entry ring per actor (≥ 1 s at 60 Hz) of `{tick, pos, CollisionBox, State, Updown, AnimCurrentFrame}`, written at the end of every server tick.
- While executing a command for a player who is in `StateAttack`, the server:
  1. computes `t = clamp(ViewTick+frac, SV.Tick - maxUnlag, SV.Tick)`;
  2. temporarily replaces each *other* actor's `Pos` and `Updown` with the history value lerped at `t`;
  3. runs the attacker's `Update`;
  4. restores the saved values.
- Hit or block outcomes are applied to the target's live state. `enterAttackedState`/`enterBlockState` only touch state, animation and speed, not `Pos`, so the restore does not clobber them.
- Eligibility checks (`State != StateAttacked/StateBlock`) read the live state by keeping `State` unrewound.
- `maxUnlag` defaults to 250 ms (15 ticks), and `sv_lagcomp` can be switched off. A debug overlay stores the last rewound boxes.
- Rewinding guard height as well as position deliberately favors the attacker, consistent with Source. The trade-off is below.

### D8. Loops and modes
- **Offline:** unchanged `runLevel` flow with `CL` authoritative.
- **Listen:** per frame:
  1. `netServerPoll`, then run the server tick accumulator: execute commands, bots step, history, snapshot every 3rd tick.
  2. `netClientPoll`, then the client tick accumulator: sample, predict, send.
  3. Interpolate remotes, then render.

  The local client talks to the in-process server over real ENet loopback (127.0.0.1), so there is one code path and loopback RTT is about 0.
- **Dedicated:** `cmd/main.go` parses `-dedicated -port -maxplayers -bots -lagcomp`. `resistor.Dedicated(cfg)` loads the level (no textures, no window) and loops with `time.Sleep` to the next tick.
- **Menu:** raygui buttons and a text box for the address. `rl.EnableCursor()` while in the menu and `rl.DisableCursor()` on entering a level.
- **Online flags:** the `Multiplayer` bool is replaced by `NetMode` (`NetOffline`, `NetListen`, `NetClient`, `NetDedicated`). Rewind input, `LE.Update` and `LE.SaveLevel` are gated on `NetOffline`.
- **Net simulation:** outgoing and incoming packets pass through a delay queue with a random drop filter, configured from the debug overlay.

## Risks / Trade-offs

- **[Float divergence between client and server builds]** Go may fuse multiply-adds on some architectures (arm64). → Positions are truncated to integers each tick, which absorbs most drift. Reconciliation corrects the rest. Players are asked to use the same build (enforced by the protocol version).
- **[Rewound guard height makes "I blocked but got hit" possible for a defender]** → Keep `maxUnlag` modest (250 ms). It is a server setting and can be retuned without a spec change. The alternative (current guard) is noted for playtesting.
- **[Per-actor hit-stop feels different online]** Other fighters keep moving during a hit-stop. → Offline is identical. Online, this is the only fair option for more than two fighters.
- **[Lagging player freezes in place (D3)]** A client whose commands stop does not advance, which Source accepts. → Timeouts remove stalled peers. Being hit still applies.
- **[Actor-to-actor interactions during prediction]** Block/attacked decisions made by the opponent's attack appear on the local fighter one RTT late. → This is inherent to authoritative netcode. Smoothing plus server-confirmed effects keep it readable.
- **[Refactor regresses single-player feel]** Moving animation timing to ticks changes the frame rounding from `PhysicsTime` floats to integer ticks. → Verify each animation offline before networking work, with a debug comparison of frame and finished flags.
- **[cgo / ENet on every build]** → Already required by the examples, and raylib already needs a C toolchain on Windows.
- **[Draw-order sort mutates `Actors`]** `slices.SortFunc(Actors, ...)` in `runLevel` reorders the simulation slice. → Sort a separate draw list, so iteration order in `updateAttack` stays stable (matters for determinism between server and prediction).

## Migration Plan

1. Land D1 (and the draw-sort fix) offline only. Verify single player plays identically.
2. Add `World` (D2) and `NetMode`, still offline.
3. Add protocol, session and snapshots with interpolation, without prediction (clients wait for the server).
4. Add prediction/reconciliation, then lag compensation, then bots, the menu and diagnostics.

Each step leaves single player working. Rollback is a revert of the step; there is no persisted data or level format change.

## Open Questions

- Tuning values: interp delay (100 ms), unlag cap (250 ms), smoothing threshold (8 px), token-bucket burst (8). Set by playtesting under the net simulator.
- Spawn point placement: derive from level geometry, or add spawn markers to the level editor later. For now, fixed X offsets along the ground, validated against collision.
