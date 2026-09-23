## Purpose

Lets a server change the rules of a match with Lua scripts - a gamemode and optional filterscripts in the manner of San Andreas Multiplayer - which players join with the plain client.

## ADDED Requirements

### Requirement: Gamemode and filterscripts
The server SHALL run at most one gamemode, named on the command line and read from `scripts/gamemodes/<name>.lua`, and any number of filterscripts, named as a comma-separated list and read from `scripts/filterscripts/<name>.lua`, both relative to the working directory, in release builds as in development builds. Each script SHALL run in a state of its own that no other script can see. Filterscripts SHALL be loaded in the order named, before the gamemode. With no gamemode and no filterscripts named, the server SHALL behave exactly as it does without scripting.

#### Scenario: A gamemode is loaded
- **WHEN** the dedicated server is started with gamemode `tdm` and `scripts/gamemodes/tdm.lua` exists
- **THEN** the script runs, its init callback is called once, and the startup log names the gamemode

#### Scenario: Filterscripts alongside it
- **WHEN** the server is started with gamemode `tdm` and filterscripts `admin,welcome`
- **THEN** `admin` is loaded, then `welcome`, then `tdm`, and a global set by one of them is not visible to the others

#### Scenario: No scripts
- **WHEN** the server is started with neither a gamemode nor filterscripts
- **THEN** the match plays exactly as without scripting

### Requirement: Scripts run on the server only
Scripts SHALL run only in the authoritative server - the dedicated server or a game hosted from the menu. A joining client or spectator SHALL need no script files and SHALL run none. Single player SHALL run no scripts.

#### Scenario: A plain client joins a scripted server
- **WHEN** a client with no `scripts` folder joins a server running a gamemode
- **THEN** it plays the match under the gamemode's rules

#### Scenario: Single player
- **WHEN** the game is started with a gamemode named and the player chooses single player
- **THEN** no script runs and time manipulation works as without scripting

### Requirement: Sandbox
A script SHALL have Lua's base, string, table, math and bit32 libraries, without any function that reads or runs files or code from a string. It SHALL NOT have the io, os, debug or package libraries. `print` SHALL write to the server log, prefixed with the script's name.

#### Scenario: No file access
- **WHEN** a script calls `io.open`, `os.execute`, `dofile` or `require`
- **THEN** the call fails as an error in that script, and nothing on the disk is read or changed

#### Scenario: Printing
- **WHEN** gamemode `tdm` calls `print("round 1")`
- **THEN** the server log shows the line with `tdm` beside it

### Requirement: Callbacks
The server SHALL call these global functions in a script, when the script defines them, with these arguments:

- `OnGameModeInit()` / `OnGameModeExit()` in the gamemode, and `OnFilterScriptInit()` / `OnFilterScriptExit()` in a filterscript, when it is loaded and when the server shuts down.
- `OnPlayerConnect(playerid)` when a fighter player is accepted, and `OnPlayerDisconnect(playerid, reason)` when one leaves, with reason `"quit"`, `"timeout"` or `"kicked"`.
- `OnFighterSpawn(fighterid, playerid)` when a fighter enters the world - at a player's connect, at a respawn, or added by a script - with `playerid` 0 for a bot.
- `OnFighterHit(fighterid, attackerid, structure)` when a hit lands, with the structure it cost.
- `OnFighterBlock(fighterid, attackerid)` when a guard or block-all stops a blow.
- `OnFighterDeath(fighterid, killerid)` when a fighter's structure breaks, with `killerid` 0 when nobody is to blame.

Every callback SHALL be called in each filterscript in load order and then in the gamemode. Callbacks for what happened during a server tick SHALL be called after that tick, in the order the events happened, and a change a callback makes SHALL hold: it is not undone by the server's lag compensation.

#### Scenario: A kill is reported
- **WHEN** player 1's fighter breaks a bot's structure
- **THEN** `OnFighterDeath` is called with the bot's fighter id and player 1's fighter id

#### Scenario: A change made in a callback holds
- **WHEN** a script sets a fighter's structure to full inside `OnFighterHit`
- **THEN** the next snapshot shows that fighter at full structure

#### Scenario: Order
- **WHEN** filterscript `admin` and gamemode `tdm` both define `OnPlayerConnect` and a player connects
- **THEN** `admin`'s is called first, then `tdm`'s

### Requirement: Player and fighter ids
A player id SHALL identify one connected fighter player for as long as it is connected, across its deaths and respawns. A fighter id SHALL identify one fighter for one life: a respawned fighter SHALL have a new id. A function given an id that no longer names a player or a living fighter SHALL do nothing and return `nil` or `false`, and SHALL NOT raise an error.

#### Scenario: A fighter id after its death
- **WHEN** a script keeps a fighter id and calls `GetFighterStructure` with it after that fighter was removed from the world
- **THEN** it gets `nil`, and the script goes on

#### Scenario: A player across a respawn
- **WHEN** a player's fighter dies and respawns
- **THEN** `GetPlayerFighter` with the same player id returns the new fighter's id

### Requirement: Script functions
A script SHALL be able to call these global functions:

- `GetTick()`: the server tick.
- `GetPlayers()`: the ids of the connected fighter players; `IsPlayerConnected(playerid)`; `GetPlayerFighter(playerid)`: its living fighter's id, or `nil`; `KickPlayer(playerid)`.
- `GetFighters()`: the ids of the fighters in the world; `IsBot(fighterid)`; `GetFighterPlayer(fighterid)`: its player's id, 0 for a bot.
- `GetFighterPos(fighterid)` returning x, y; `SetFighterPos(fighterid, x, y)`, which puts the fighter there standing still, and refuses - returning `false` and leaving it where it was - a spot where its body would be inside the level's solid ground.
- `GetFighterStructure` / `SetFighterStructure`, `GetFighterEnergy` / `SetFighterEnergy`, each set clamped to the fighter's range; setting structure to 0 or less breaks the fighter as `KillFighter` does.
- `GetFighterTeam` / `SetFighterTeam`, team 0 being on nobody's side.
- `KillFighter(fighterid)`: the fighter dies as if its structure broke, with nobody to blame.
- `AddBot()` / `AddBot(x, y)`: a new bot, returning its fighter id; `RemoveBot(fighterid)`.
- `SendPlayerMessage(playerid, text)` and `SendMessageToAll(text)`.
- `SetTimer(func, ms, repeating)` returning a timer id, and `KillTimer(timerid)`.

A change a function makes SHALL reach every client by the snapshots, as the same change made by the game would.

#### Scenario: Teams
- **WHEN** a script puts two fighters on team 1 with `SetFighterTeam`
- **THEN** their swings pass through one another, on the server and on every client

#### Scenario: Structure is clamped
- **WHEN** a script sets a fighter's structure above the maximum
- **THEN** the fighter has the maximum

#### Scenario: Not into a wall
- **WHEN** a script moves a fighter to a spot inside the floor
- **THEN** `SetFighterPos` returns `false` and the fighter stays where it was

#### Scenario: A bot added by a script
- **WHEN** a script calls `AddBot()`
- **THEN** a bot spawns at a free spawn position, `OnFighterSpawn` is called for it with player id 0, and it fights as any server bot does

### Requirement: Timers
A timer SHALL call its function once after the given time, or, when repeating, every that long until it is killed, measured in server ticks and rounded up to at least one tick. A timer killed from inside its own function SHALL NOT fire again. Timers SHALL stop when the server shuts down.

#### Scenario: A repeating timer
- **WHEN** a script sets a repeating 1000 ms timer
- **THEN** its function is called once every 60 server ticks until `KillTimer` is called with its id

### Requirement: Script errors are contained
An error raised in a callback or timer SHALL stop that call only: it SHALL be written to the server log with the script's name and the error, the other scripts' callbacks for the same event SHALL still be called, and the match SHALL go on. A single call that runs past a fixed instruction budget SHALL be stopped and reported the same way. A gamemode or filterscript that is missing, does not compile, or raises an error while its main chunk runs SHALL keep the server from starting, with the reason shown.

#### Scenario: An error in a callback
- **WHEN** a gamemode's `OnFighterDeath` indexes a nil value
- **THEN** the server log shows the error with the gamemode's name, the fighter still dies and respawns, and the next death calls `OnFighterDeath` again

#### Scenario: An endless loop
- **WHEN** a filterscript's `OnPlayerConnect` loops forever
- **THEN** the call is stopped within a fraction of a second, the error is logged, the gamemode's `OnPlayerConnect` still runs, and the server keeps its tick rate

#### Scenario: A gamemode that does not compile
- **WHEN** the dedicated server is started with a gamemode that has a syntax error
- **THEN** it prints the file, line and error and exits without listening
