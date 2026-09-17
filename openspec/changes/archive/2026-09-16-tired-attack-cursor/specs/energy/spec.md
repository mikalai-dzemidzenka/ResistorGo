## ADDED Requirements

### Requirement: The attack crosshair shows energy
The player's attack crosshair — the one that trails the mouse and chooses the guard and attack height — SHALL be coloured by the local fighter's energy: the full green it has always had when the pool is full, blending continuously towards a pale grey-green as the pool empties, and that pale grey-green at zero. The blend SHALL follow the energy fraction directly, with no step at any threshold. The crosshair SHALL stay fully opaque at every energy. The mouse crosshair SHALL NOT change. With no local fighter in the world, the attack crosshair SHALL be drawn as at a full pool.

#### Scenario: Full pool
- **WHEN** the local fighter's energy is full
- **THEN** the attack crosshair is the same green it was before this change

#### Scenario: Draining
- **WHEN** the local fighter swings repeatedly and its energy falls
- **THEN** the attack crosshair grows paler with each swing, and at half the pool it is about halfway between the green and the pale grey-green

#### Scenario: Empty
- **WHEN** the local fighter's energy is 0
- **THEN** the attack crosshair is pale grey-green and still clearly visible over both lit and unlit parts of the level

#### Scenario: Refilling
- **WHEN** the local fighter stops swinging and its energy regenerates
- **THEN** the attack crosshair turns gradually back to green as the pool refills

#### Scenario: No fighter
- **WHEN** the local fighter has died and not yet respawned
- **THEN** the attack crosshair is the full green

### Requirement: A tired fighter chooses its height more slowly
The attack crosshair SHALL follow the mouse at a speed set by the local fighter's energy fraction: at a full pool, as quickly as it did before this change at 60 frames per second — settling on a new mouse position in about 0.07 seconds — and at an empty pool, several times slower, settling in about 0.35 to 0.4 seconds, with the speed falling continuously in between. Settling time SHALL NOT depend on the frame rate. Because the attack crosshair chooses the guard and attack height, a fighter below full energy SHALL take correspondingly longer to change height. Blocking SHALL remain available at any energy: once the crosshair has reached a height, a block at that height SHALL resolve exactly as it would at full energy. Bots SHALL be unaffected. Only the chosen height SHALL travel to the server, as it always has.

#### Scenario: Full pool feels as before
- **WHEN** a fighter at full energy snaps the mouse from its feet to above its head
- **THEN** its guard goes to high as quickly as it did before this change

#### Scenario: Empty pool is sluggish
- **WHEN** a fighter at 0 energy snaps the mouse from its feet to above its head
- **THEN** the attack crosshair visibly trails behind the mouse, and the guard reaches high noticeably later than it would at full energy

#### Scenario: Late to the guard
- **WHEN** a fighter at low energy moves the mouse from a low guard to a high one just before an opponent's high attack lands
- **THEN** the attack is met by whatever height the attack crosshair has reached at that moment, not by the height of the mouse

#### Scenario: An empty fighter still blocks once there
- **WHEN** a fighter at 0 energy holds the attack crosshair at the attacker's height and is attacked
- **THEN** the attack is blocked as it would be at full energy

#### Scenario: No step at full
- **WHEN** a fighter's energy goes from full to one point short of full
- **THEN** the attack crosshair's speed changes imperceptibly rather than dropping to a slower setting

#### Scenario: Any frame rate
- **WHEN** the same fighter at the same energy moves the mouse the same way at 30 and at 144 frames per second
- **THEN** the attack crosshair takes the same time to settle

#### Scenario: Bots are not slowed
- **WHEN** a bot with an empty pool changes its guard
- **THEN** it does so as quickly as a bot with a full pool

#### Scenario: Nothing new on the wire
- **WHEN** a tired player plays against a server
- **THEN** the server receives the same kind of input it always has — the height the attack crosshair chose — and neither the protocol nor a server that predates this change needs to know about the slowdown
