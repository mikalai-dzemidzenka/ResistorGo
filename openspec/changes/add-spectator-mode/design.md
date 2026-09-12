## Context

See proposal.md for motivation, `specs/multiplayer/spectator/spec.md` for the new behavior, and the session and replication deltas for the changed requirements. The current code (the enet multiplayer change) assumes every accepted client owns a fighter:

- **Server (`net_server.go`).**
  - A peer counts as accepted only when `netPlayer.Actor != nil`.
  - `humans()` counts those peers, and `Step` sends snapshots only to them.
  - `Poll` drops any peer whose `LastRecv` is older than 10 s. Fighters refresh it with 60 command packets a second; a spectator would send nothing after its Hello.
  - ENet peer capacity is `MaxPlayers+4`.
- **Client (`net_client.go`).** "In the match" means `c.W.Local != nil`. That condition gates entering the level (in `reconcile`), the connect-timeout vs lost-timeout choice, fixed-tick command sending, and which menu screen `fail` returns to. `applyInterp` skips `c.ActorID`.
- **Protocol (`net_protocol.go`).** `Hello{Proto u16, LevelHash u64}` is a fixed 11-byte message, and `ProtocolVersion` is 1.
- **Level loop (`game.go`).** The camera targets `CL.Local`; `Cam.Update` (camera.go) handles wheel zoom and smooths the target at 1000 px/s. The OS cursor is disabled in the level, so mouse movement is read as a delta. `Cur.DrawGUI` draws the attack cursor.
- **Menu (`main_menu.go`).** The Join screen has Connect and Back. The Host screen has port, max players and bots.

## Goals / Non-Goals

**Goals:**
- Spectators reuse the existing snapshot and interpolation path unchanged. A spectator is "a client with no local fighter".
- The camera logic is pure (no raylib input reads inside it), so it can be unit tested.

**Non-Goals:**
- Any change to fighter simulation, prediction or lag compensation.
- Hiding information from players who spectate a match they also play in, and any spectator delay.

## Decisions

### D1. Role travels in Hello; protocol version 2
- `Hello` gains `Role u8` (`RoleFighter=0`, `RoleSpectator=1`), and `ProtocolVersion` becomes 2.
- `decodeHello` reads the role only if a byte remains. A version-1 Hello (11 bytes) therefore still decodes, and the server rejects it with "version mismatch" instead of silently dropping it, so an old client isn't left to time out.
- New values:
  - `RejectSpectatorsFull` shows as "spectators full".
  - `MsgKeepalive` (see D2).
- `Welcome` keeps its layout. `ActorID = 0` means "no fighter", since actor ids start at 1.
- **Alternative considered:** a separate `MsgSpectate` handshake. Rejected, because it would mean two handshake paths to validate.

### D2. Server: accepted-without-fighter peers
- `netPlayer` gets `Role` and `Accepted`:
  - `humans()` counts accepted fighters;
  - a new `spectators()` counts accepted spectators;
  - snapshot recipients are all accepted peers, and a spectator's `Ack` is 0.
- **Handshake order:** version, then level, then cap. The cap is `MaxPlayers` for fighters (as now) or `MaxSpectators` for spectators.
- **Spectator accept:** no `makeRoom`, no spawn, no bot refill on leave. `Welcome{ActorID: 0}` is sent, and the connect/reject/disconnect log lines say "spectator".
- **Commands from spectators:** `MsgCmds` from a spectator is dropped before decoding.
- **Keepalive:** a spectator client sends `MsgKeepalive` (1 byte, unreliable channel) once a second. Any received packet already refreshes `LastRecv`, so the server only ignores the message body. One lost keepalive can't trip the 10 s timeout.
- **Capacity and flags:**
  - `ServerConfig.MaxSpectators` defaults to 4.
  - ENet peer capacity becomes `MaxPlayers + MaxSpectators + 4`, which keeps the room for rejecting extra clients.
  - The new flag is `-maxspectators`.
- **Alternative considered:** exempt spectators from the app timeout and rely on ENet's own peer timeout. Rejected, because ENet's timeout varies between 5 and 30 s with RTT, and disconnect detection would then differ by role.

### D3. Client: "ready" instead of "has a local fighter"
- `Client.Spectator bool`, set by `startClient(addr, spectator)`, which Hello uses.
- `c.ready()` is `c.W.Local != nil || (c.Spectator && c.HasSnap)`. It replaces the `Local != nil` checks for the timeouts and for `fail`'s menu screen.
- **Entering the level:**
  - A fighter enters on its first own record, as now.
  - A spectator enters on its first snapshot, from `onSnapshot`.
- **Ticks:** a spectator runs no fixed ticks (nothing to predict, no commands). Its keepalive goes through `NetSim.send`, so simulated latency and loss apply.
- `applyInterp` needs no change: `ActorID` is 0, so every fighter is a remote, interpolated actor.

### D4. Spectator camera (`spectator.go`)
```go
type SpecCam struct {
    Follow   bool
    FollowID uint16 // 0 = follow the first fighter that becomes visible
    Pos      rl.Vector2
}
```
- **Pure helpers:**
  - `nextFighter(ids []uint16, cur uint16, dir int) (uint16, bool)`: ids are sorted ascending; it wraps; `cur` missing means it starts from the first (or last) id.
  - `panStep(pos, dir rl.Vector2, zoom, dt float32) rl.Vector2`: world speed is `SpecPanSpeed / zoom`, with `SpecPanSpeed` = 400 screen px/s, which gives the "same on-screen speed at any zoom" requirement.
  - `dragStep(pos, mouseDelta rl.Vector2, zoom float32) rl.Vector2`: `pos − delta/zoom`.
- **`(*SpecCam).Update(in SpecInput, w *World, zoom, dt)`** applies the mode rules from the spec:
  - Q/E cycle fighters and enter follow mode.
  - Space or any pan/drag input switches to free mode at the current `Pos`.
  - A missing followed fighter switches to the next one, or to free mode.
  - `FollowID == 0` picks the first visible fighter.
  - `SpecInput` is read from raylib by one small function in game.go.
- **Wiring into `Cam`:**
  - In follow mode the result goes through `Cam.SetTarget` (the fighter position + 24, the same smoothing as the player camera).
  - In free mode it sets both `Cam.Target` and `Cam.Camera2D.Target`, so panning and dragging aren't lagged by the 1000 px/s smoothing.
- **Zoom:** `Cam.Update` keeps handling zoom, and `SpecCam` reads `Cam.Zoom`.
- **Alternative considered:** extending `Camera` itself with modes. Rejected, because it would mix player-camera behavior with spectator-only state.

### D5. Level loop and HUD
- **In `runLevel`:** when `Cli != nil && Cli.Spectator`, skip `latchInput`, run `Spec.Update` instead of targeting `CL.Local`, and don't draw `Cur.DrawGUI`.
- **HUD:** drawn in screen space after the palette pass, always (not only in debug), at text size 20: "SPECTATING", the mode ("free" / "following #<actor id>"), and a key-hint line.
- **Keys:**
  - Esc leaves the match as it does now.
  - The debug keys (`` ` ``, F5–F8) are unchanged.
  - Q, E, Space, WASD and the right mouse button have no other meaning for spectators.

### D6. Menu
- **Join screen:** Connect, Join as spectator, Back. Both joins set `connectingFrom = menuJoin`.
- **Host screen:** a "Max spectators" spinner (0–8, default 4) is added above Start.
- 0 spectators is a valid cap; every spectator is then rejected with "spectators full".

## Risks / Trade-offs

- **[Version 1 builds can no longer join]** → The handshake reports "version mismatch" to old clients (D1), and the bump is required anyway because Hello changed.
- **[Spectators see the live match without delay]** A player could watch from a second client to scout. → This is accepted for now; a spectator delay is a listed non-goal.
- **[Snapshot bandwidth grows with spectators]** About 0.7 KB × 20/s per spectator at 16 fighters. → The default cap is 4 and is configurable.
- **[Mouse-delta drag with a disabled cursor]** The raylib delta can jump on the first frame after the window regains focus. → Clamp a single frame's delta to one screen size (tuning).

## Migration Plan

Ship server and client together (protocol 2). There is no data migration. Rollback is a revert; an older build simply rejects newer clients with "version mismatch".

## Open Questions

- Tuning: pan speed (400 screen px/s) and the first-frame drag clamp. Set by playtesting.
