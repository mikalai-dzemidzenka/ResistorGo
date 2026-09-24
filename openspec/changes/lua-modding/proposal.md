## Why

Every match plays by the rules compiled into the binary: free-for-all against
bots, respawn, repeat. San Andreas Multiplayer showed how far a game goes when
the server runs scripts that decide the rules - deathmatch, teams, races,
rounds, whole game modes nobody on the original team imagined - while players
join with the plain client. Resistor already has the shape SA-MP had: one
authoritative server that owns every fighter, and clients that only send input
and draw snapshots. Scripts on that server can change the game for everyone
without touching a client.

## What Changes

- The server - dedicated or hosted from the menu - can run Lua scripts, using
  the pure-Go VM `github.com/Shopify/go-lua` (Lua 5.2). Single player and
  clients run no scripts.
- As in SA-MP, one **gamemode** is picked with `-gamemode <name>` from
  `scripts/gamemodes/<name>.lua`, and any number of **filterscripts** ride
  alongside it with `-filterscripts a,b` from `scripts/filterscripts/`. Each
  script runs in its own Lua state. With no `-gamemode` the server behaves
  exactly as it does today.
- Scripts are told what happens through global callbacks named after SA-MP's:
  init and exit, a player connecting and leaving, a fighter spawning, being
  hit, blocking and dying, and timers they set.
- Scripts act through a set of functions: list players and fighters, read and
  set a fighter's position, structure, energy and team, kill a fighter, add and
  remove bots, set timers, send text, and log to the server console.
- A new server-to-client text message: a script sends a line to one player or
  to everyone, and the client shows it in a short message feed over the level.
  Players cannot type yet.
- Scripts are sandboxed: no file, OS, debug or module access; a callback that
  runs too long or errors is stopped and logged, and the match goes on. A
  gamemode that fails to load stops the server from starting.

## Capabilities

### New Capabilities
- `modding`: loading gamemodes and filterscripts, the sandbox, the callbacks
  scripts receive, the functions they may call, and how script errors and
  runaway scripts are contained.
- `multiplayer/server-messages`: text lines the server sends to one client or
  all of them, and the message feed a client shows them in.
- `teams`: a red or blue line under the fighters a gamemode puts on teams 2
  and 3, and no attack outline on an ally. Without it an ally, whom a swing
  passes through, looks like an enemy that cannot be hurt.

### Modified Capabilities
- `multiplayer/session`: the dedicated server and a hosted game accept
  `-gamemode` and `-filterscripts`; a gamemode that fails to load keeps the
  server from starting, with the error shown.
- `multiplayer/replication`: a snapshot carries each fighter's team. A client
  today puts every remote fighter on the bots' team, which is harmless while
  teams never change, and wrong once a script sets them.

## Impact

- New dependency `github.com/Shopify/go-lua`. It has no coroutines and no Lua
  patterns (`string.gsub`, `match` and `gmatch` are missing); the API is
  designed around that.
- New code: a script host (`script.go`: loading, sandbox, dispatch, timers) and
  its API bindings (`script_api.go`).
- `world.go` / `actor.go`: an authoritative world with scripts records game
  events (spawn, hit, block, death) as they happen; the server hands them to
  the scripts at the end of its tick, outside the lag-compensated rewind.
- `net_server.go`: connect and disconnect callbacks, the tick hook, the text
  message; `net_protocol.go`: a new reliable message, a team byte in each
  fighter's snapshot record (46 to 47 bytes; 16 fighters still fit a packet)
  and a protocol version bump; `net_client.go` and the level HUD: the message
  feed.
- `cmd/main.go`: the two flags; they apply to a hosted game as well.
- No change to single player or time manipulation.
