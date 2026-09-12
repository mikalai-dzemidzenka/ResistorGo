## 1. Make the simulation independent of rendering (offline only, D1)

- [ ] 1.1 Add `Ticks`/`AnimStartTick` to `Actor`; make `SetAnim`/`SetAnimFrame` stamp `AnimStartTick` instead of `AnimPhysTimeStart`; add `advanceAnim()` at the end of `Actor.Update` that computes `AnimCurrentFrame` and `AnimFinished` (logic moved out of `DrawAnimOnce`/`DrawAnim`); make draw functions read-only. Verify: `go build ./...`, then play single player and check that walk, run, attack chains (all heights), block, hit-stun, jump, climb and turnover finish and transition as before (debug text "Frame, Finished").
- [x] 1.2 Introduce an explicit `animTable []string` and `animID(path) uint8`; make `LoadTextures` load from it. Verify: game starts with no missing-texture panic, and every anim path used in actor.go is in the table (grep each `resources/threefour/` literal).
- [ ] 1.3 Extend `Input` with `Updown`, `FaceLeft`, `ViewTick`, `ViewFrac`; add `sampleInput(a)` that computes guard height from `Cur.AtkPos`/`Cam` and facing from the mouse side; make `updateUpdown`/`MouseOnLeft` read only `a.Input`. Verify: single player guard height follows the attack cursor and facing follows the mouse exactly as before.
- [ ] 1.4 Move bot randomness (guard changes, attack rolls) into `botInput(a)`, which targets the nearest other actor instead of `Pl`. Verify: the offline bot still approaches, changes guard and attacks.
- [ ] 1.5 Replace `Cam.ZoomSpeed`/`Cam.Offset` mutations in actor code with `a.Events` flags (attack start, hit, blocked) consumed by presentation code in `runLevel` for the viewed actor only. Verify: offline, attack zoom and hit shake still happen when the local fighter attacks or is hit, and actor code no longer references `Cam`.
- [ ] 1.6 Replace global `PauseTime` hit-stop with per-actor `FreezeTicks` on attacker and defender (6 ticks); remove `PauseTime`. Verify: offline hit-stop looks the same (both fighters freeze ~0.1 s).
- [ ] 1.7 Sort a separate draw slice in `runLevel` instead of sorting `Actors` in place. Verify: draw order is unchanged visually, and the `Actors` order stays stable across frames (debug print once).

## 2. World and net mode scaffolding (D2, D8)

- [ ] 2.1 Add `World{Actors, Tick, Authority, History}`, the `SV`/`CL` globals and `Actor.W`; route `updateAttack` and bot targeting through `a.W.Actors`; offline creates one authoritative `CL` with `CL.Local` replacing `Pl`. Verify: `go build ./...` and single player unchanged.
- [ ] 2.2 Replace the `Multiplayer` bool with `NetMode` (`NetOffline`, `NetListen`, `NetClient`, `NetDedicated`); gate `UpdateRewind`, `LE.Update` and `LE.SaveLevel` on `NetOffline`. Verify: offline rewind/editor still work; with NetMode forced to `NetClient`, R does nothing and `level_1.json` mtime is unchanged after exit.
- [x] 2.3 Skip hit resolution in `updateAttack` when `!a.W.Authority` (the attack animation still plays). Verify: with `CL.Authority=false` forced offline, swings animate but never hit the bot.

## 3. Protocol (D4)

- [x] 3.1 Create `net_protocol.go` with message types, protocol version, `Hello`/`Welcome`/`Reject`/`Bye`/`Cmds`/`Snapshot` encode/decode, the full actor record, and a length check on every decoder. Verify: `net_protocol_test.go` round-trips every message, rejects truncated and unknown-type buffers without panicking, and asserts that a 16-actor snapshot is ≤ 1200 bytes; `go test ./...` passes.
- [x] 3.2 Add `levelHash()` (FNV-1a of the `level_1.json` bytes). Verify: unit test gives the same hash for the same bytes and a different hash after a one-byte change.

## 4. Server and session (session spec, replication spec)

- [x] 4.1 Create `net_server.go`: ENet host with 2 channels on the configured port, a connect/receive/disconnect loop on the main goroutine, packet data copied before `Destroy`, handshake validation (version, level hash, human cap) with `Reject` reasons, and `Welcome` with player/actor ids. Verify: run the dedicated server (5.1) and connect a client with a mismatched protocol version or edited level: the client shows the matching reason.
- [x] 4.2 Add a per-player command queue ordered by `Seq` with duplicate/old dropping, and a token bucket (+1/tick, burst 8) driving `Update` of that player's actor once per executed command. Verify: unit test feeds duplicated, reordered and 2×-rate command streams and asserts the execution count and order.
- [x] 4.3 Add the server tick loop: bots step each tick, snapshot every 3rd tick per client with that client's `ackSeq`, sent unreliable on channel 0. Verify: the client debug overlay shows ~20 snapshots/s.
- [x] 4.4 Add spawn placement (fixed X offsets on the ground, skipping positions overlapping another actor's collision box) and actor removal on disconnect/timeout. Verify: with 3 clients joining one after another, no fighters overlap at spawn; closing one client removes its fighter on the others within 1 s.
- [x] 4.5 Add bot slot management: keep the configured bot count, remove a bot when a human needs a slot, refill when a human leaves. Verify: host with max 4, bots 3: a second human join succeeds and the bot count drops to 2; after they leave it returns to 3.
- [x] 4.6 Add the host-quit `Bye{server closed}` and ENet peer timeout handling. Verify: quitting the host sends clients to the menu with "server closed"; killing the server process yields "connection lost" on clients within 10 s.

## 5. Modes and entry points (D8)

- [x] 5.1 Parse `-dedicated -port -maxplayers -bots -lagcomp` in `cmd/main.go`; add `resistor.Dedicated(cfg)`, which loads the level without a window or textures, loops at 60 Hz with sleep, and logs connects/disconnects/rejections. Verify: `go run ./cmd -dedicated` prints the listening port and no window opens.
- [ ] 5.2 Add listen mode: `runLevel` drives the server accumulator (`SV`), then the client accumulator (`CL`), with the local client connecting to 127.0.0.1 over ENet. Verify: Host Game spawns the host fighter; a second `go run ./cmd` instance can Join 127.0.0.1 and both see each other.
- [x] 5.3 Create `net_client.go`: connect with a 5 s timeout, handle `Welcome`/`Reject`/`Bye`, discard old snapshots, and return to the menu with a reason on failure or loss. Verify: joining an address with no server shows "connection timed out" after ~5 s.

## 6. Interpolation (interpolation spec, D6)

- [ ] 6.1 Create `net_interp.go`: 32-snapshot buffer, smoothed `estServerTick` (±5 %, never backwards), `renderTick = est - interpDelay`, and per-remote-actor lerp of position and `Ticks` with discrete fields from the older record; derive the anim frame with the shared anim function; round positions for drawing. Verify: unit test with synthetic snapshots checks bracketing, rounding and the discrete-field choice; in-game, a remote walking fighter shows every walk frame.
- [x] 6.2 Add extrapolation up to 250 ms then hold, teleport detection, and spawn/remove resolved at `renderTick`. Verify: unit tests for the 150 ms gap (keeps moving), the 2 s gap (stops after 250 ms) and a teleport record pair (snaps).
- [x] 6.3 Write `renderTick` into each outgoing command's `ViewTick`/`ViewFrac`. Verify: server debug log shows ViewTick ≈ serverTick − 6 on loopback.

## 7. Prediction and reconciliation (prediction spec, D5)

- [x] 7.1 Add the 128-entry command ring; each client tick, sample → store → send `Cmds` (newest + 2 previous) → `Local.Update` in the non-authoritative `CL`. Verify: under 200 ms artificial lag (8.2), movement and attack start respond on the next tick.
- [x] 7.2 On snapshot: copy the server record into the local actor, drop acked commands, replay the rest with events suppressed, and count corrections. Verify: walking on flat ground with 150 ms lag keeps the correction counter at 0; getting hit by a bot shows hit-stun after reconciliation.
- [ ] 7.3 Add `CorrectionOffset` smoothing: ≤ 8 px decays over ~100 ms, larger snaps; applied to the drawn position only. Verify: with 100 ms lag and 10 % loss, the local fighter never visibly jitters during normal walking.
- [ ] 7.4 Drive presentation effects (hit flash, shake, zoom) for remote fighters from interpolated state transitions, and for the local fighter from non-replay events and server-confirmed state. Verify: swinging at an opponent that is out of range on the server never flashes it; replaying an attack does not re-zoom.

## 8. Lag compensation and diagnostics (lag-compensation spec, D7)

- [x] 8.1 Create `net_lagcomp.go`: a per-actor 64-tick history written at the end of each server tick, and `rewindOthers(t)`/`restore()` swapping `Pos`/`Updown` (not `State`); wrap attacker command execution with the clamped view time (maxUnlag 250 ms) when `-lagcomp` is on. Verify: unit test that rewind-then-restore leaves positions identical, and that view ticks older than the window and in the future are clamped.
- [ ] 8.2 Add the network simulator: a delay queue (0–500 ms) and a random drop (0–50 %) on client send and receive, configurable from the debug overlay. Verify: setting 200 ms raises the displayed RTT by ~200 ms.
- [ ] 8.3 Add the debug net overlay (RTT from `Peer.GetRoundTripTime`, interp delay, snapshots/s, corrections) and, on a listen server, drawing of the last rewound hitboxes. Verify: the overlay values update live with debug mode on.
- [ ] 8.4 Verify the lag-comp scenarios end-to-end: with 150 ms simulated lag and lag comp on, hitting an opponent who stepped back registers a hit; with `-lagcomp=false` the same swing misses; high-guard-switch-after-view is resolved against the old guard.

## 9. Main menu (session spec)

- [ ] 9.1 Add raygui menu entries (Single Player, Host Game, Join Game, Quit) drawn over the wireframe menu; `rl.EnableCursor()` in the menu and `rl.DisableCursor()` on entering a level; default `Stage` becomes `StageMainMenu`. Verify: the game boots into the menu, the entries are clickable, and Single Player starts the offline level as before.
- [ ] 9.2 Add the Host screen (port, max players, bots, Start/Back) with a bind-failure error, and the Join screen (address text box accepting `host[:port]`, Connect/Cancel, connecting state, error line). Verify: an invalid address shows an error without connecting; a busy port shows an error naming it.

## 10. End-to-end verification

- [ ] 10.1 Play a match on one machine: a listen host plus 2 joined clients plus 1 bot, with 150 ms lag and 5 % loss on one client. Verify every session, replication, prediction and interpolation scenario in the specs by hand, and note any tuning changes (interp delay, unlag cap, smoothing threshold) in design.md Open Questions.
- [ ] 10.2 Play a match on a dedicated server with 2 clients on separate machines on a LAN. Verify join/leave, combat and bot refill; confirm single player is still unchanged after all steps.
