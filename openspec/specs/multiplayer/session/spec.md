# multiplayer/session Specification

## Purpose

Lets players start, host, join and leave networked matches, either on a listen server run by a player or on a headless dedicated server. Covers slots, spawning, bots and features that are unavailable while online.

## Requirements

### Requirement: Main menu entries
The game SHALL boot into the main menu. The menu SHALL offer Single Player, Host Game, Join Game, Options and Quit, selectable with the mouse. The mouse cursor SHALL be visible and usable while the menu is shown.

#### Scenario: Boot shows the menu
- **WHEN** the game is started without the dedicated-server flag
- **THEN** the animated main menu is displayed with the entries Single Player, Host Game, Join Game, Options and Quit

#### Scenario: Single player unchanged
- **WHEN** the user selects Single Player
- **THEN** the level starts offline with the local fighter and one bot, and gameplay, rewind and the level editor behave as before this change

#### Scenario: Quit
- **WHEN** the user selects Quit
- **THEN** the game closes

### Requirement: Listen server hosting
Host Game SHALL start an authoritative server inside the game process. The server SHALL listen on a UDP port (default 27015). The host SHALL join it as a regular player. The host screen SHALL let the user set the port, the maximum player count and the number of bots before starting.

#### Scenario: Host starts a match
- **WHEN** the user confirms Host Game with default settings
- **THEN** the server listens on UDP port 27015, the host's fighter spawns in the level, and other clients can connect

#### Scenario: Port unavailable
- **WHEN** the user confirms Host Game and the chosen port cannot be bound
- **THEN** the game stays in the menu and shows an error naming the port

#### Scenario: Host leaves
- **WHEN** the hosting player quits the match
- **THEN** every connected client is disconnected with the reason "server closed" and returns to the main menu

### Requirement: Dedicated server
When started with the dedicated-server command-line flag, the game SHALL run the same authoritative server without creating a window or loading textures. It SHALL accept optional flags for port, maximum players and bot count. It SHALL log connects, disconnects and rejections to standard output, and run until the process is terminated.

#### Scenario: Headless start
- **WHEN** the game is started with the dedicated-server flag
- **THEN** no window is opened, the server listens on the configured port, and a startup line with the port is printed

#### Scenario: All players are clients
- **WHEN** two players join a dedicated server
- **THEN** both are remote clients with identical behavior, and the server has no local fighter

### Requirement: Joining a server
Join Game SHALL let the user enter an address as `host` or `host:port` (default port 27015) and connect. While connecting, the game SHALL show a connecting state that can be cancelled. The attempt SHALL time out after 5 seconds without an answer, returning to the menu with an error.

#### Scenario: Successful join
- **WHEN** the user enters the address of a running server and confirms
- **THEN** the client connects, is assigned a player, loads the level and spawns its fighter

#### Scenario: No server
- **WHEN** the user joins an address where no server answers within 5 seconds
- **THEN** the game returns to the Join screen showing "connection timed out"

#### Scenario: Invalid address
- **WHEN** the user enters an address that cannot be parsed or resolved
- **THEN** no connection is attempted and an error is shown

### Requirement: Handshake validation
On connect, the client SHALL send its protocol version and a fingerprint of its level data. The server SHALL reject the client, with a reason shown in the client's menu, if the protocol version differs, the level fingerprint differs, or the server is full of human players.

#### Scenario: Version mismatch
- **WHEN** a client with a different protocol version connects
- **THEN** it is rejected and the client displays "version mismatch"

#### Scenario: Level mismatch
- **WHEN** a client whose level file differs from the server's connects
- **THEN** it is rejected and the client displays "level mismatch"

#### Scenario: Server full
- **WHEN** a client connects while the number of human players equals the maximum
- **THEN** it is rejected and the client displays "server full"

### Requirement: Player slots and spawning
Every accepted player SHALL receive a unique player id and a fighter. The fighter SHALL spawn at a spawn position not occupied by another fighter's collision box. All other clients SHALL see the new fighter no later than the second snapshot after it spawns. A player whose fighter dies SHALL keep its slot and player id while it has no fighter, and SHALL be given a new one when it respawns; its slot SHALL NOT be offered to a joining player in the meantime.

#### Scenario: Second player joins
- **WHEN** a second player joins a match with one player in it
- **THEN** both clients show two fighters, and the fighters do not overlap at spawn

#### Scenario: Slot survives a death
- **WHEN** a player's fighter dies and is removed from the world
- **THEN** the player stays connected and keeps its slot and player id, and a client connecting during that gap is told the server is full if every slot is taken

#### Scenario: Respawn takes a free position
- **WHEN** a dead player's fighter respawns while another fighter is standing on the spawn position it used before
- **THEN** it spawns at a position not occupied by another fighter's collision box, and every client sees it no later than the second snapshot after it spawns

### Requirement: Leaving and connection loss
When a player disconnects or times out, the server SHALL remove that player's fighter, and all remaining clients SHALL stop displaying it. A client that loses its connection to the server SHALL return to the main menu with the message "connection lost" within 10 seconds of the loss.

#### Scenario: Player quits
- **WHEN** a client quits the match
- **THEN** its fighter disappears on all other clients

#### Scenario: Network cable pulled
- **WHEN** a client stops receiving anything from the server
- **THEN** within 10 seconds it returns to the main menu showing "connection lost"

### Requirement: Server-side bots
The server SHALL maintain the configured number of bot fighters. Bots SHALL be simulated only on the server, target the nearest other fighter, and appear to clients exactly like remote players. Bots SHALL occupy player slots. When a human joins and all slots are taken, one bot SHALL be removed to make room; when a human leaves, bots SHALL refill up to the configured count.

#### Scenario: Bots fill the match
- **WHEN** a server is started with 2 bots and one human joins
- **THEN** that client sees its own fighter and two bot fighters

#### Scenario: Human replaces bot
- **WHEN** max players is 4, 3 bots are configured, one human is in the match, and a second human joins
- **THEN** the second human is accepted and one bot is removed

### Requirement: Offline-only features are disabled online
While in a networked match (host or client), time manipulation, the level editor and writing the level file SHALL be disabled. In a development build, debug overlays SHALL remain available; a release build has none.

#### Scenario: Rewind key online
- **WHEN** a player in a networked match presses T
- **THEN** nothing is paused or rewound and gameplay continues normally

#### Scenario: Level not saved
- **WHEN** a player exits the game after a networked match
- **THEN** the level file on disk is unchanged

#### Scenario: Overlay online in a development build
- **WHEN** a player of a development build toggles the debug overlay in a networked match
- **THEN** the overlay is shown and the match is unaffected

### Requirement: Network diagnostics
In debug mode, a networked client SHALL display round-trip time, interpolation delay, snapshot receive rate and the number of prediction corrections. The client SHALL also offer a network simulation setting that adds configurable artificial latency (0–500 ms) and packet loss (0–50 %) to its traffic.

#### Scenario: Overlay shows ping
- **WHEN** debug mode is on during a networked match
- **THEN** the current round-trip time in milliseconds is displayed

#### Scenario: Simulated lag
- **WHEN** artificial latency is set to 200 ms against a local listen server
- **THEN** the displayed round-trip time rises by about 200 ms, and the local fighter still responds to input immediately
