## Context

See proposal.md for why. The pieces this builds on:

- `Server.Step` (`net_server.go`) is the authoritative tick: respawns, then
  each player's queued commands (`runCmd`, which rewinds the other fighters for
  lag compensation around an attacking fighter's `Update` and restores them
  after), then the bots, then history and snapshots. A hosted game runs the
  same `Server` in-process and joins it over loopback, so the host is a client
  like any other.
- Hits, blocks and deaths are decided inside `Actor.Update`: `takeHit`,
  `confirmBlock`, `enterDeathState`; `World.Kill` / `reap` / `updateRespawns`
  handle removal and return, and already call the server's
  `OnFighterRemoved` / `OnFighterRespawned` hooks.
- Single player is an authoritative `World` too, with time manipulation that
  restores actors by value. A Lua state cannot be rewound with it, which is
  why single player runs no scripts (specs/modding).
- `github.com/Shopify/go-lua` is Lua 5.2 in pure Go: `NewState`, per-library
  openers (`BaseOpen`, `StringOpen`, `TableOpen`, `MathOpen`, `Bit32Open`, ...),
  `LoadBuffer`, `ProtectedCall`, and `SetDebugHook` with a count mask. It has
  no coroutines and no Lua patterns.

## Goals / Non-Goals

**Goals:**
- A script cannot crash, stall or corrupt the server: every call into Lua is
  protected and budgeted.
- A script sees the world only between ticks, never half-way through one.
- Scripting costs nothing when no script is loaded.

**Non-Goals:**
- Player chat and `/commands`, client-side scripts, sending scripts to clients.
- Callbacks that veto what the game decided (cancelling a hit, refusing a
  connect); every callback here is a notification.
- Calls between scripts (SA-MP's `CallRemoteFunction`), `require`, hot reload.
- Player names, and anything the level editor or level file would have to
  carry (script-placed objects, spawn points).

## Decisions

### One Lua state per script, dispatched in load order

A `script` holds its name, kind (gamemode / filterscript) and `*lua.State`;
the `scriptHost` holds them in load order, filterscripts first, and the timers.
`host.call(name, args...)` looks up the global in each state and, when it is a
function, calls it protected. Separate states are what SA-MP's separate AMX
instances gave: a filterscript cannot clobber a gamemode's globals.
Alternative: one shared state with scripts as modules - rejected, one script's
global leak or error would reach every other.

### Events are queued during the tick and delivered after it

`World` gets a nil-able `Script *scriptHost`. `takeHit`, `confirmBlock`,
`enterDeathState` and the spawn paths append a small `scriptEvent{kind,
fighter, other ID, amount}` to it when it is set, and do nothing else. At the
end of `Server.Step`, after the bots and before history and snapshots, the
host delivers the queue in order, then fires due timers. Two reasons:
- `runCmd` rewinds other fighters for lag compensation around an attack, and
  a hit is decided inside that window. A callback run there that moved or
  healed a rewound fighter would be overwritten by `restore`. Delivered after
  the tick, a change holds (specs/modding, "a change made in a callback holds").
- Scripts calling `KillFighter` or `AddBot` mid-`Update` would change
  `w.Actors` while the server iterates it.

Events carry actor IDs, not pointers, and functions resolve IDs with
`World.ByID` at call time, so a fighter reaped between the event and the call
simply is not found. The single player world never sets `Script`, so its cost
there is one nil check.

Connect and disconnect are called directly from `accept` / `drop`, which run
in `Poll`, outside any tick.

### Death is reported once, at `enterDeathState`

`enterDeathState` is where a fighter's structure breaks, with the attacker in
hand (nil for block-all or a script kill → killerid 0). The later removal is
the world's bookkeeping; reporting there would lose the killer and come a
death animation late.

### Sandbox by opening libraries one by one

Instead of `lua.OpenLibraries`, the host requires `_G` (base), `string`,
`table`, `math` and `bit32` individually, then sets `dofile`, `loadfile`,
`load` and `loadstring` to nil and replaces `print` with one that writes to the
server log as `[script <name>] ...`. `io`, `os`, `debug` and `package` are
never opened, so `require` does not exist. Scripts are read from disk by Go
(`os.ReadFile`) and compiled with `LoadBuffer`, named by path so errors carry
file and line.

### An instruction budget per call

Each state gets a `SetDebugHook` with the count mask, every 10 000
instructions, set once when the state is made; every protected call starts the
script's count at zero, and the hook raises a Lua error once a call passes
5 000 000 instructions (about 40 ms in go-lua, measured), which unwinds to the
`ProtectedCall` like any other error. Alternative: a wall-clock deadline checked in the hook - rejected as
nondeterministic under load; the budget is a count.

### Errors are logged per script and the server goes on

A failed call logs `[script <name>] <callback>: <error>` and the host moves to
the next script. A load failure (missing file, compile error, error in the
main chunk) is returned from `NewServer`, so `Dedicated` prints it and exits
with status 1, and the Host screen shows it as it shows a port it cannot bind.
`OnGameModeInit` / `OnFilterScriptInit` run as part of loading, but an error
in them is logged, not fatal - it is a callback like the rest.

### Timers in ticks

`SetTimer(fn, ms, repeating)` stores the function in a table in the state's
registry under the timer's id - go-lua has no `luaL_ref` - and a timer `{id,
script, every, due, repeating}` in the host; `every = max(1, ceil(ms*60/1000))`. Due timers fire after the event
queue, in id order. `KillTimer` marks it dead, so killing a timer from its own
callback, or another due in the same tick, is safe. SA-MP timers are in
milliseconds, which is why the API keeps ms.

### Ids

A player id is the existing `netPlayer.ID`: stable while connected. A fighter
id is `Actor.ID`, which the world never reuses and which changes at respawn
(`updateRespawns` spawns a new actor). `GetPlayerFighter` goes through the
peer's current `Actor`. Bots have player id 0, as `Actor.Owner` already has.

### Functions write the same fields the game does

Setters clamp and write `Structure`, `Energy`, `Team`, `Pos` directly.
`SetFighterPos` refuses a spot where the fighter's collision box overlaps a
solid, and otherwise moves it standing still and runs `checkCollision` for its
footing. `checkCollision` lands a fighter and catches ledges but never pushes
one out of a wall, so a fighter placed inside one would stay stuck there.
Structure set to 0 or less is a break, as `KillFighter`: a fighter standing on
no structure is a state the game never has. `KillFighter` calls
`enterDeathState(nil)`, and the fighter falls and is reaped the usual way.
`AddBot` is `fillBots`' spawn plus an entry in `Server.Bots`; `RemoveBot` is `World.Remove` of a bot with no pending respawn.
`KickPlayer` is `drop` with reason `"kicked"`, then an ENet disconnect.

### Team in the snapshot

`ActorRec` gains a `Team uint8` (record 46 → 47 bytes) and `apply` sets it.
Today a client builds every remote fighter with `NewActor(pos, false)` and so
puts it on `BotTeam`, which only works because nothing ever changes a team.
Client prediction reads teams in crowding and shoving, so a team a script sets
has to reach it.

### Text message

`MsgText` on `ChanReliable`: `[type][len u8][utf-8 bytes]`, capped at 128
bytes, cut back to a rune boundary on the server. The client's decoder rejects
a length that overruns the packet or a body that is not valid UTF-8. Runes
outside the font's printable ASCII are drawn as `?` (raylib's default font).
The feed is a small ring of `{text, until}` on the `Client`, drawn in
`drawLevel`'s window-space pass under the spectator HUD's text style, six lines
up from the bottom-left, and cleared with the client. `ProtocolVersion` goes
from 8 to 9 for the new message and record size.

### Flags

`-gamemode` and `-filterscripts` set package vars `Gamemode` and
`Filterscripts` (like `LevelPath`), which `NewServer` reads, so the dedicated
server and a hosted game both pick them up without the Host screen changing.

## Risks / Trade-offs

- [go-lua is incomplete: no coroutines, no patterns] → The API needs neither:
  timers stand in for waiting, and there is no text input to parse. The
  proposal names the gap so modders are not surprised.
- [go-lua is slower than C Lua, and the hook costs a little more] → Callbacks
  are per event, not per tick; a tick hook (`OnServerTick`) is deliberately
  left out of v1. The budget caps the worst case.
- [A callback's change arrives one tick "late"] → It lands before the tick's
  snapshot, so no client ever sees the in-between state.
- [Team byte and a new message break old clients] → The protocol version bump
  turns that into the existing "version mismatch" rejection.
- [Scripts load from the working directory] → The same place `settings.json`
  lives; the startup line names the full path it loaded from.

## Migration Plan

None needed: with no flags the server is unchanged, and old clients are
turned away by the version check instead of misreading the new record.

## Open Questions

- The instruction budget (5 000 000, about 40 ms here) may want tuning on a
  slower dedicated server machine.
