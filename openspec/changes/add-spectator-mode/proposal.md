## Why

Online matches can only be joined as a fighter, so nobody can watch a game without taking a slot or joining the fight. A full server can't be watched at all. The multiplayer change deliberately left spectators out of scope (see `openspec/changes/archive/2026-09-12-add-enet-multiplayer/proposal.md`, Non-goals). The pieces spectating needs already exist: the server sends full snapshots, and the client can interpolate every fighter. Spectating mostly adds a way to join without a fighter, plus a camera that isn't tied to one.

## What Changes

- **Join as spectator.** The Join screen gets a "Join as spectator" action next to Connect. A spectator gets snapshots like any client but has no fighter, sends no input commands, and cannot affect the match.
- **Separate spectator cap.** Spectators don't use player slots. The server has its own maximum number of spectators, 4 by default, set on the Host screen and by a dedicated-server flag. A server full of fighters can still be watched. When the spectator cap is reached, new spectators are rejected with "spectators full".
- **Spectator camera** with two modes:
  - *Free*: pan with WASD or the arrow keys, drag with the right mouse button, zoom with the mouse wheel.
  - *Follow*: Q and E step to the previous or next fighter, and the camera tracks that fighter the way it tracks the local fighter today. Space or any pan input returns to free mode from the current view. If the followed fighter leaves, the camera goes to the next fighter, or back to free mode if none are left.
- **Spectator HUD**: shows "SPECTATING", the current camera mode, which fighter is followed, and the key hints.
- **Protocol version 2.** The handshake carries the joining role (fighter or spectator). Clients on version 1 are still told "version mismatch".
- **Keepalive.** Spectators send a small periodic message so the server's inactivity timeout doesn't drop them, since they send no input commands.

## Capabilities

### New Capabilities
- `multiplayer/spectator`: joining without a fighter, receiving and displaying the match, spectator camera modes and controls, the spectator HUD, and the rule that spectators cannot influence the match.

### Modified Capabilities
- `multiplayer/session`:
  - *Joining a server* adds the spectator join action.
  - *Handshake validation* adds the role, the separate spectator cap and the "spectators full" reject. "Server full" now applies only to fighters.
  - *Listen server hosting* adds the max-spectators setting on the Host screen.
  - *Dedicated server* adds the max-spectators flag.
- `multiplayer/replication`: *Input commands* applies only to clients that control a fighter. Spectators send no commands.

## Impact

- **Code**
  - `net_protocol.go`: Hello role, protocol version 2, keepalive message, "spectators full" reject reason.
  - `net_server.go`: spectator acceptance and cap, snapshots to spectators, keepalive counted as activity, ENet peer count includes spectators.
  - `net_client.go`: spectator mode, a "ready" state that doesn't need a local fighter, no command sending, keepalive.
  - `main_menu.go`: Join and Host screen controls.
  - `game.go`: spectator camera and HUD in the level loop.
  - `camera.go` or a new `spectator.go`: camera mode logic.
  - `cmd/main.go`: `-maxspectators` flag.
- **Compatibility:** protocol version 2 is not compatible with version 1 builds. Mismatched builds get a clear "version mismatch" rejection.
- **Not affected:** single player, the fighter simulation, prediction, lag compensation and the level format.

### Non-goals
- Switching between fighter and spectator during a match.
- Spectating in single player (offline free camera).
- Demo recording or playback, a spectator delay or a spectator chat.
