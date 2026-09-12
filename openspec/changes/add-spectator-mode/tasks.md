## 1. Protocol (D1)

- [x] 1.1 In `net_protocol.go`: add `Role` to `Hello` (`RoleFighter`, `RoleSpectator`), bump `ProtocolVersion` to 2, add `MsgKeepalive` and `RejectSpectatorsFull` ("spectators full"); make `decodeHello` read the role only when a byte remains. Verify: `net_protocol_test.go` round-trips a spectator Hello, decodes an 11-byte version-1 Hello without error (role = fighter), and `rejectText(RejectSpectatorsFull)` is "spectators full"; `go test ./...` passes.

## 2. Server (D2)

- [x] 2.1 Add `Role`/`Accepted` to `netPlayer`, `spectators()`, and `ServerConfig.MaxSpectators` (default 4); raise the ENet peer capacity to `MaxPlayers + MaxSpectators + 4`; make snapshot recipients all accepted peers (spectators get `Ack` 0). Verify: `go build ./...`, and the existing integration test still passes.
- [x] 2.2 In the Hello handling: check the cap by role, accept spectators without spawning, `makeRoom` or bot refill, send `Welcome{ActorID: 0}`, and log "spectator" on connect, reject and disconnect; drop `MsgCmds` from spectators and ignore `MsgKeepalive`. Verify: an integration test with raw peers where max players is 1 and max spectators is 1. A second fighter gets "server full". A spectator on the full server is accepted, no bot is removed, and it receives snapshots. A second spectator gets "spectators full". Forged commands from the spectator leave every fighter's state unchanged. The spectator leaving leaves the bot count and free slots unchanged.
- [x] 2.3 Add the `-maxspectators` flag in `cmd/main.go`. Verify: `resistor.exe -dedicated -maxspectators 1` prints the startup line; a second raw spectator is rejected and the rejection is logged.

## 3. Client (D3)

- [x] 3.1 Add `Client.Spectator` and `startClient(addr, spectator)` (Hello carries the role); add `c.ready()` and use it for the timeouts and `fail`'s menu screen; enter the level on the first snapshot for spectators; skip fixed ticks and command sending for spectators; send `MsgKeepalive` once a second through `NetSim`. Verify: an integration test with a real spectator `Client` against a server with 1 human (raw peer) and 2 bots. It enters, sees 3 fighters in `CL.Actors`, and sends no `MsgCmds`: the server's spectator queue is empty and fighter actors are untouched. With `PeerTimeout` lowered for the test, it is still connected after 3× the timeout.
- [x] 3.2 Update `startListen` and other callers for the new `startClient` signature. Verify: `go build ./...` and the existing listen-mode integration test pass.

## 4. Spectator camera (D4)

- [x] 4.1 Create `spectator.go` with `SpecCam`, `SpecInput`, `nextFighter`, `panStep`, `dragStep` and `(*SpecCam).Update` implementing the free/follow rules from the spectator spec. Verify with a `spectator_test.go` unit test covering:
  - Q/E wrap over three ids in order;
  - pan distance scaled by zoom gives the same on-screen distance at zoom 1 and 5;
  - a 100 px drag moves the view 100 screen px;
  - Space or pan input leaves follow mode without moving `Pos`;
  - a followed id that disappears switches to the next id, and to free mode when none are left;
  - `FollowID == 0` picks the first id that appears.
- [ ] 4.2 In `runLevel`, when spectating: read `SpecInput` from the keyboard, the right mouse button, and the mouse delta (clamped to one screen per frame); skip `latchInput`; drive `Cam` from `Spec.Update` (follow mode goes through `SetTarget`, free mode sets the target directly); keep wheel zoom in `Cam.Update`; don't draw the attack cursor. Reset `Spec` when a spectator enters the level. Verify by playtest: pan, drag, zoom, Q/E, Space and follow-on-leave behave as the spec describes.

## 5. HUD and menu (D5, D6)

- [ ] 5.1 Draw the spectator HUD in screen space after the palette pass, always visible: "SPECTATING", the mode ("free" / "following #<id>") and the key hints. Verify by playtest: the HUD is visible with debug mode on and off and updates when the mode changes.
- [ ] 5.2 Join screen: add "Join as spectator" (calls `startClient(addr, true)`, then the connecting screen). Host screen: add a "Max spectators" spinner (0–8, default 4) passed into `ServerConfig`. Verify by playtest: joining a listen host as a spectator from a second instance works, and a host with max spectators 0 shows "spectators full" to the spectator.

## 6. End-to-end verification

- [ ] 6.1 Play on one machine: a listen host with 1 bot, a second instance joins as a fighter, and a third as a spectator. Check every spectator scenario and the changed session scenarios by hand. Watch for 60 s without input to confirm the spectator stays connected, and quit the host to confirm the spectator sees "server closed". Note any tuning (pan speed, drag clamp) in design.md Open Questions.
