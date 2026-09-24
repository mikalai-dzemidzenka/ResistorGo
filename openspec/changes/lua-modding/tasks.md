## 1. Script host

- [x] 1.1 Add `github.com/Shopify/go-lua` to `go.mod`; verify `go build ./...`
      and `build.ps1` still produce a working exe
- [x] 1.2 `script.go`: a `scriptHost` loading filterscripts then a gamemode
      from `scripts/filterscripts/` and `scripts/gamemodes/`, one sandboxed
      `lua.State` each (base, string, table, math, bit32; no `dofile`,
      `loadfile`, `load`, `loadstring`; `print` to the server log with the
      script's name); verify with tests on scripts in a temp dir that load
      order is kept, globals do not leak between scripts, and `io`, `os`,
      `debug`, `require` and `dofile` are unavailable
- [x] 1.3 Protected `call(name, args...)` over every script in order, with the
      instruction budget armed per call; verify with tests that an error in one
      script is logged with its name and the next script is still called, and
      that `while true do end` is stopped and logged within the budget
- [x] 1.4 Load failures - missing file, compile error, error in the main
      chunk - return an error naming the file (and line when there is one);
      verify each with a test
- [x] 1.5 Timers: `SetTimer` / `KillTimer`, due in ticks, fired in id order;
      verify with tests a one-shot fires once, a repeating 1000 ms timer every
      60 ticks, and a timer killed in its own callback does not fire again

## 2. Game events

- [x] 2.1 `World.Script` and a `scriptEvent` queue; `takeHit`, `confirmBlock`,
      `enterDeathState` and the spawn paths (connect, respawn, `AddBot`)
      append to it only when `Script` is set; verify with tests that each
      event is queued once with the right ids and cost, and that a world
      without `Script` queues nothing
- [x] 2.2 Deliver the queue and the due timers at the end of `Server.Step`,
      before history and snapshots; verify with a test that a script healing
      a fighter in `OnFighterHit` during a lag-compensated attack still has it
      healed in the next snapshot
- [x] 2.3 `OnPlayerConnect` from `accept`, `OnPlayerDisconnect` from `drop`
      with `"quit"`, `"timeout"` or `"kicked"`; init and exit callbacks at load
      and at `Server.Close`; verify with a server test that counts them

## 3. Script functions

- [x] 3.1 `script_api.go`: `GetTick`, `GetPlayers`, `IsPlayerConnected`,
      `GetPlayerFighter`, `GetFighters`, `IsBot`, `GetFighterPlayer`; verify
      with tests, including `nil` for a fighter id after its fighter was
      reaped and a new fighter id for a player after a respawn
- [x] 3.2 Position, structure, energy and team getters and setters, clamped;
      a move into solid ground refused; verify with tests of clamping and of
      a move into a wall
- [x] 3.3 `KillFighter`, `AddBot`, `RemoveBot`, `KickPlayer`; verify with tests
      that a killed fighter dies with killer 0 and respawns, an added bot
      fights and gets `OnFighterSpawn` with player 0, and a kicked player is
      disconnected with reason `"kicked"`

## 4. Network

- [x] 4.1 `Team` in `ActorRec` (46 → 47 bytes) and `ProtocolVersion` 9; verify
      `TestActorRecSize`, `TestSnapshotFitsPacket`, `TestRecApplyRoundTrip`
      and a new test that a team crosses the wire
- [x] 4.2 `MsgText` encode/decode, capped at 128 bytes on a rune boundary,
      malformed bodies rejected; `SendPlayerMessage` / `SendMessageToAll` on
      the reliable channel; verify with protocol tests and with the malformed
      packet test extended to the new type
- [x] 4.3 Client message feed: six lines, ten seconds each, window space,
      shown with the HUD hidden, cleared on leaving; verify the feed's ring
      and expiry with a test, then by eye (5.3)

## 5. Launching

- [x] 5.1 `-gamemode` and `-filterscripts` flags into package vars read by
      `NewServer`; `Dedicated` exits 1 with the load error; the Host screen
      shows it; the startup line names the gamemode; verify a server test
      with a failing gamemode gets the error and does not listen
- [x] 5.2 An example gamemode in `scripts/gamemodes/tdm.lua` (two teams,
      a message on each kill, a score every minute by timer) and a
      filterscript `scripts/filterscripts/welcome.lua`; verify both load in a
      test
- [x] 5.3 Run `go test ./...` and `go test -tags release ./...`; build with
      `build.ps1`

## 6. In the game

- [x] 6.1 Dedicated server with `-gamemode tdm -filterscripts welcome` and two
      clients: the welcome line shows, the teams hold, kill messages and the
      score appear in the feed (two clients on a hosted game, which runs the
      same server; the dedicated one checked headless with bots)
- [x] 6.2 Host Game from a game started with `-gamemode tdm`: the host plays
      under it; started with a broken gamemode: the Host screen shows the error
- [x] 6.3 Single player with `-gamemode tdm`: no script runs, time
      manipulation works
- [x] 6.4 A gamemode with an endless loop in `OnFighterDeath`: the server logs
      it and keeps playing smoothly (dedicated, 4 bots on two teams: each death
      logs "stopped: ran past 5000000 instructions", the fighter respawns and
      timers keep their period; a runaway call is stopped in ~20 ms)

## 7. Teams on screen

- [x] 7.1 A red line under a fighter on team 2 and a blue one under team 3,
      the body's width, on the ground under the feet, over the lighting with
      the bars and hidden with the HUD; no line for no team or the bots' team;
      verify with tests of the colours (palette entries), when it shows and
      where it goes
- [x] 7.2 The attack outline skips an ally; verify in the outline exclusions
      test that an ally is not marked, an enemy is, and two fighters with no
      team still are
- [x] 7.3 TDM leaves bots neutral on the bots' team: they fight both sides,
      and breaking one scores nothing; verify in `TestExampleScripts` that two
      players are split, two bots stay put, and only a kill of the other side
      scores
- [x] 7.4 In the game: a TDM host and a second client with bots - each sees
      its own side's colour and the other's, the bots have no line and both
      sides can hurt them, and a teammate is never outlined
