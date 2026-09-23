## MODIFIED Requirements

### Requirement: Listen server hosting
Host Game SHALL start an authoritative server inside the game process. The server SHALL listen on a UDP port (default 27015). The host SHALL join it as a regular player. The host screen SHALL let the user set the port, the maximum player count and the number of bots before starting. The server SHALL run the gamemode and filterscripts named on the game's command line, if any.

#### Scenario: Host starts a match
- **WHEN** the user confirms Host Game with default settings
- **THEN** the server listens on UDP port 27015, the host's fighter spawns in the level, and other clients can connect

#### Scenario: Port unavailable
- **WHEN** the user confirms Host Game and the chosen port cannot be bound
- **THEN** the game stays in the menu and shows an error naming the port

#### Scenario: Host leaves
- **WHEN** the hosting player quits the match
- **THEN** every connected client is disconnected with the reason "server closed" and returns to the main menu

#### Scenario: Hosting a gamemode
- **WHEN** the game was started with gamemode `tdm` and the user confirms Host Game
- **THEN** the hosted server runs `tdm`

#### Scenario: A gamemode that fails to load
- **WHEN** the game was started with a gamemode that does not compile and the user confirms Host Game
- **THEN** the game stays in the menu and shows an error naming the script

### Requirement: Dedicated server
When started with the dedicated-server command-line flag, the game SHALL run the same authoritative server without creating a window or loading textures. It SHALL accept optional flags for port, maximum players, bot count, gamemode and filterscripts. It SHALL log connects, disconnects and rejections to standard output, and run until the process is terminated. A gamemode or filterscript that fails to load SHALL make it print the reason and exit without listening.

#### Scenario: Headless start
- **WHEN** the game is started with the dedicated-server flag
- **THEN** no window is opened, the server listens on the configured port, and a startup line with the port is printed

#### Scenario: All players are clients
- **WHEN** two players join a dedicated server
- **THEN** both are remote clients with identical behavior, and the server has no local fighter

#### Scenario: Scripts named on the command line
- **WHEN** the dedicated server is started with `-gamemode tdm -filterscripts admin`
- **THEN** it loads `admin` and `tdm` before listening, and the startup line names the gamemode
