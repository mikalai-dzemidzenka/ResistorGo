## MODIFIED Requirements

### Requirement: Listen server hosting
Host Game SHALL start an authoritative server inside the game process. The server SHALL listen on a UDP port (default 27015). The host SHALL join it as a regular player. The host screen SHALL let the user set the port, the maximum player count, the number of bots and the maximum number of spectators (default 4) before starting. The server SHALL run the gamemode and filterscripts named on the game's command line, if any.

#### Scenario: Host starts a match
- **WHEN** the user confirms Host Game with default settings
- **THEN** the server listens on UDP port 27015, the host's fighter spawns in the level, and other clients can connect as fighters or as up to 4 spectators

#### Scenario: Port unavailable
- **WHEN** the user confirms Host Game and the chosen port cannot be bound
- **THEN** the game stays in the menu and shows an error naming the port

#### Scenario: Host leaves
- **WHEN** the hosting player quits the match
- **THEN** every connected client, spectators included, is disconnected with the reason "server closed" and returns to the main menu

#### Scenario: Hosting a gamemode
- **WHEN** the game was started with gamemode `tdm` and the user confirms Host Game
- **THEN** the hosted server runs `tdm`

#### Scenario: A gamemode that fails to load
- **WHEN** the game was started with a gamemode that does not compile and the user confirms Host Game
- **THEN** the game stays in the menu and shows an error naming the script

### Requirement: Dedicated server
When started with the dedicated-server command-line flag, the game SHALL run the same authoritative server without creating a window or loading textures. It SHALL accept optional flags for port, maximum players, bot count, maximum spectators, gamemode and filterscripts. It SHALL log connects, disconnects and rejections to standard output, distinguishing spectators from players, and run until the process is terminated. A gamemode or filterscript that fails to load SHALL make it print the reason and exit without listening.

#### Scenario: Headless start
- **WHEN** the game is started with the dedicated-server flag
- **THEN** no window is opened, the server listens on the configured port, and a startup line with the port is printed

#### Scenario: All players are clients
- **WHEN** two players join a dedicated server
- **THEN** both are remote clients with identical behavior, and the server has no local fighter

#### Scenario: Scripts named on the command line
- **WHEN** the dedicated server is started with `-gamemode tdm -filterscripts admin`
- **THEN** it loads `admin` and `tdm` before listening, and the startup line names the gamemode

#### Scenario: Spectator cap flag
- **WHEN** the dedicated server is started with a maximum of 1 spectator and a second spectator connects
- **THEN** the second spectator is rejected, and the rejection is logged

### Requirement: Joining a server
Join Game SHALL let the user enter an address as `host` or `host:port` (default port 27015) and connect, either as a fighter (Connect) or as a spectator (Join as spectator). While connecting, the game SHALL show a connecting state that can be cancelled. The attempt SHALL time out after 5 seconds without an answer, returning to the menu with an error.

#### Scenario: Successful join
- **WHEN** the user enters the address of a running server and confirms with Connect
- **THEN** the client connects, is assigned a player, loads the level and spawns its fighter

#### Scenario: Successful spectator join
- **WHEN** the user enters the address of a running server and chooses Join as spectator
- **THEN** the client connects, loads the level and shows the match without spawning a fighter

#### Scenario: No server
- **WHEN** the user joins an address where no server answers within 5 seconds
- **THEN** the game returns to the Join screen showing "connection timed out"

#### Scenario: Invalid address
- **WHEN** the user enters an address that cannot be parsed or resolved
- **THEN** no connection is attempted and an error is shown

### Requirement: Handshake validation
On connect, the client SHALL send its protocol version, a fingerprint of its level data and whether it joins as a fighter or a spectator. The server SHALL reject the client, with a reason shown in the client's menu, if the protocol version differs, or the level fingerprint differs. It SHALL also reject a fighter when the server is full of human players, and a spectator when the maximum number of spectators is connected. Spectators SHALL NOT count toward the player maximum, and fighters SHALL NOT count toward the spectator maximum. A client from a build with an older protocol version SHALL be told "version mismatch", not left to time out.

#### Scenario: Version mismatch
- **WHEN** a client with a different protocol version connects
- **THEN** it is rejected and the client displays "version mismatch"

#### Scenario: Level mismatch
- **WHEN** a client whose level file differs from the server's connects
- **THEN** it is rejected and the client displays "level mismatch"

#### Scenario: Server full
- **WHEN** a client joining as a fighter connects while the number of human players equals the maximum
- **THEN** it is rejected and the client displays "server full"

#### Scenario: Full server can be watched
- **WHEN** a client joins as a spectator while the number of human players equals the maximum and fewer spectators than the maximum are connected
- **THEN** it is accepted as a spectator

#### Scenario: Spectators full
- **WHEN** a client joins as a spectator while the maximum number of spectators is connected
- **THEN** it is rejected and the client displays "spectators full"
