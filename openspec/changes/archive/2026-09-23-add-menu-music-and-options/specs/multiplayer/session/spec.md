## MODIFIED Requirements

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
