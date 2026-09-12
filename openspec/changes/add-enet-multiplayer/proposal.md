## Why

Resistor is single-player only: `Multiplayer` in game.go is a flag with no implementation behind it (`Actor.updateMultiplayerInput` is empty). As a side effect the flag is `true` by default, so the enemy actor currently receives no input at all. The only networking code is `examples/example9`, a throwaway prototype. Its physics is a separate copy of the game's, it predicts once per rendered frame instead of per fixed tick, and it smooths remote players with a lerp instead of a timed buffer. None of it can be reused for the real `Actor` state machine.

A melee game with three guard heights is timing-critical, so we want the Source-engine netcode model: an authoritative server, client-side prediction with reconciliation, buffered interpolation of other players, and server-side lag compensation for hits. Players get responsive controls and fair hits while the server stays the only source of truth.

## What Changes

- **Authoritative server over ENet** (go-enet). The server runs the existing 60 Hz fixed-step simulation, consumes client input commands, and broadcasts world snapshots at 20 Hz on an unreliable channel. Reliable messages (welcome, join/leave, level check) go on a separate channel.
- **Two server modes, one server code path.**
  - *Listen server*: the host plays in the same process as a local client.
  - *Dedicated server*: headless, no window, started with a command-line flag (a headless process has no menu).
- **Client-side prediction and reconciliation** for the local fighter. Each input tick is simulated immediately and stored. When a snapshot acknowledges a command, the client rewinds to the server state and replays the unacknowledged commands. Small errors are smoothed; large ones snap.
- **Entity interpolation** for all other actors. The client renders them about 100 ms in the past, interpolating between buffered snapshots, with a short capped extrapolation on packet loss.
- **Lag compensation.** The server keeps a short history of every actor's position, collision box, state and animation frame. When resolving a player's attack it rewinds the other actors to the time that player saw (latency + interpolation delay, capped).
- **N players plus server-side bots.** Any number of human players up to a server cap; the server can fill slots with bots (the existing `updateBotInput`, now targeting the nearest opponent).
- **Main menu gains interactive entries**: Single Player, Host Game, Join Game (address field), Quit. These use raygui over the existing 3D wireframe menu. The game now boots into the menu instead of straight into the level.
- **Required simulation refactors** (behavior-preserving in single player):
  - Animation frame and `AnimFinished` advance in the fixed step. Today `DrawAnimOnce` computes them, so gameplay depends on rendering.
  - `Input` becomes a self-contained command: movement, jump, attack, guard height (`Updown`), and facing. `Actor.Update` stops reading the mouse, `Cur` or `Cam`.
  - Camera shake, attack zoom and hit-stop become local presentation effects driven by events, not global mutations inside `Actor.Update`.
- **Disabled in multiplayer:** time rewind (R), the level editor, and saving `level_1.json` on exit (clients must not diverge from the server level).
- **Net debug overlay** (debug mode): ping, interpolation delay, snapshot rate, prediction error count, and optionally drawn server/lag-compensation hitboxes. A configurable artificial latency/loss simulator for testing on one machine.

## Capabilities

### New Capabilities
- `multiplayer/session`: hosting (listen/dedicated), joining from the main menu, connection lifecycle (connect, welcome, level check, timeout, disconnect), player slots, spawning, and server-side bots.
- `multiplayer/replication`: server tick, client input commands (with redundancy), snapshot contents and rate, tick/clock synchronisation, and the channel/reliability rules.
- `multiplayer/prediction`: local-player prediction, command history, reconciliation against acknowledged snapshots, error smoothing, and handling of predicted combat events.
- `multiplayer/interpolation`: snapshot buffering, rendering remote actors in the past, interpolation of position/state/animation, and bounded extrapolation.
- `multiplayer/lag-compensation`: server-side actor history and rewind for hit resolution, with a latency cap.

### Modified Capabilities
None. `openspec/specs/` is empty; there are no existing specs to delta.

## Impact

- **Code**
  - `actor.go`: `Input` extended; animation timing moved into `Update`; camera/cursor/global-pause side effects removed from simulation; `Net` field used; `updateMultiplayerInput` replaced.
  - `game.go`: net mode globals; the level loop splits into server step, client predict step and render.
  - `main_menu.go`: menu items and join UI.
  - `rewind.go` / `level_editor.go`: gated to single player.
  - `cmd/main.go`: a flag for dedicated mode.
- **New files** (package `resistor`, flat layout): `net_protocol.go` (message encode/decode), `net_server.go`, `net_client.go`, `net_interp.go`, `net_lagcomp.go`.
- **Dependencies:** go-enet is already in go.mod. It is cgo and links the bundled `enet` static lib on Windows, so a C toolchain (already needed for raylib) is required for every build.
- **Network:** UDP, default port 27015, configurable.
- **Not affected:** shaders, palette, lights and level file format.
- **`examples/example9`** stays as a prototype. Its hand-written binary encoding style is the reference for `net_protocol.go`.

### Non-goals
- NAT traversal, matchmaking, server browser, encryption and authentication.
- Bit-exact cross-platform determinism. Reconciliation corrects drift instead; we rely on both sides running the same build.
- Delta-compressed snapshots (full snapshots first; delta encoding is a possible follow-up).
- Voice/chat and spectator mode.
