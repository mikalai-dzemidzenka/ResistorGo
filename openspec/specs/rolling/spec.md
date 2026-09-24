# rolling Specification

## Purpose
How a fighter makes a charged roll: it ducks while the roll button is held, the hold winds up how far and how high the roll goes, the player sees the path before letting go, and the fighter rolls along that path and recovers on landing.

## Requirements

### Requirement: Holding the roll button ducks
A fighter that is standing or running on the ground and presses the roll button (the right mouse button for the local player) SHALL stop moving and duck. It SHALL play the duck animation once and then hold its last frame, and SHALL stay ducked for as long as the button is held. A ducking fighter SHALL NOT walk, run, jump, swing or climb down. It SHALL turn to face the side of the cursor, and it SHALL roll toward that side.

A fighter in any other state SHALL NOT duck. That includes swinging, blocking, stunned, airborne, climbing or already rolling. If the button is still held when that state ends and the fighter is standing or running, it SHALL duck then. The button is read as held, not as a press.

#### Scenario: Ducking from a stand
- **WHEN** a standing fighter with enough energy presses and holds the roll button
- **THEN** it plays the duck animation, holds its last frame, and does not move while the button stays held

#### Scenario: Ducking from a run
- **WHEN** a running fighter presses the roll button
- **THEN** it stops and ducks on that tick

#### Scenario: Not in the air
- **WHEN** a fighter that is jumping or falling presses the roll button
- **THEN** it does not duck in the air, and it ducks on the tick it lands back into standing or running if the button is still held

#### Scenario: Turning while charging
- **WHEN** a ducking fighter's cursor crosses to its other side
- **THEN** the fighter turns to face the cursor

### Requirement: The hold charges the roll
How long the duck has been held SHALL set how strongly the roll launches. A longer hold SHALL launch it both higher and farther. The minimum SHALL already be a fast, low roll that carries the fighter clear of an enemy standing next to it, so a click is an escape. The charge SHALL grow from that minimum at the first tick to a maximum after about one second, and SHALL NOT grow past that maximum however long the button is held. Every machine simulating the fighter, whether server or predicting client, SHALL arrive at the same charge for the same hold.

#### Scenario: A tap still rolls
- **WHEN** a fighter presses and releases the roll button on the next tick
- **THEN** it rolls, at the minimum charge

#### Scenario: A click escapes
- **WHEN** a fighter with an enemy standing 20 px on each side clicks the roll button, holding it for about a tenth of a second
- **THEN** it rolls low and fast past the enemy it faces, through that enemy's body, and comes to rest at least 48 px beyond it

#### Scenario: Longer is higher and farther
- **WHEN** two fighters on flat ground roll, one after holding for 0.25 s and one after holding for 0.75 s
- **THEN** the second roll reaches a higher peak and lands farther away

#### Scenario: The charge caps
- **WHEN** a fighter holds the roll button for 3 s and releases it
- **THEN** it rolls exactly as it would after the maximum hold, no higher and no farther

### Requirement: The path is drawn while charging
While the local player's fighter is ducking, the game SHALL draw the arc that the roll would follow if the button were released on that tick. The arc SHALL be computed with the same launch speed, gravity, fall speed limit and time step as the roll itself, and SHALL end where it first meets a solid. It SHALL change as the charge grows and as the fighter turns. It SHALL be drawn over the lighting, so it is never shaded, and SHALL NOT be drawn for any other fighter or when the local fighter is not ducking.

#### Scenario: The arc is where the roll goes
- **WHEN** the local player releases the roll button on flat, open ground
- **THEN** the fighter lands where the arc drawn on the previous tick ended, give or take one pixel

#### Scenario: The arc grows with the hold
- **WHEN** the local player keeps holding the roll button
- **THEN** the drawn arc gets higher and longer until the charge reaches its cap, then stays the same

#### Scenario: Only while ducking
- **WHEN** the local player's fighter is doing anything other than ducking
- **THEN** no arc is drawn

### Requirement: Releasing the button rolls
When the button is released, a ducking fighter SHALL leave the ground on that tick with the speed its charge gives, toward the side it faces. It SHALL play the roll animation in a loop until it lands. While rolling, the fighter SHALL follow its arc. Steering SHALL NOT change its horizontal speed and friction SHALL NOT slow it. It SHALL NOT catch a ledge. It SHALL NOT enter the fall. A wall SHALL stop its horizontal travel, as it stops any other movement, and the roll SHALL carry on down from there.

#### Scenario: Off on release
- **WHEN** a ducking fighter releases the roll button
- **THEN** it leaves the ground on that tick, toward the side it faces, playing the roll animation

#### Scenario: Keys do not steer a roll
- **WHEN** a rolling fighter holds the key away from the way it is rolling
- **THEN** its horizontal speed does not change until it lands

#### Scenario: Past the apex is still a roll
- **WHEN** a rolling fighter starts to come down
- **THEN** it keeps playing the roll animation and does not start falling

### Requirement: Landing recovers from the roll
A rolling fighter that reaches the ground SHALL play the roll-recover animation once. During the recovery it SHALL NOT act. Any speed it lands with SHALL run out under friction. When the recovery finishes, the fighter SHALL go back to running if it is still moving fast enough to run, and to standing otherwise, following the same rule a landing from a fall uses.

#### Scenario: Recovering on landing
- **WHEN** a rolling fighter touches the ground
- **THEN** it plays the roll-recover animation once and then stands

#### Scenario: Nothing during recovery
- **WHEN** a recovering fighter presses attack or jump
- **THEN** neither happens

### Requirement: A roll costs as much as a run-jump
A roll SHALL cost the same energy as a run-jump, spent on the tick the fighter leaves the ground. Spending it SHALL restart the wait before energy refills, as spending on a swing does. A fighter without that much energy SHALL NOT duck. The press SHALL be refused: the fighter SHALL stay in the state it was in, nothing SHALL be spent, and the same refusal a swing raises SHALL be raised, so the empty bar flashes. A duck that ends without a roll, for example because a hit interrupts it, SHALL cost nothing.

#### Scenario: Paid at launch
- **WHEN** a fighter at full energy ducks, holds, and releases
- **THEN** its energy is unchanged while ducking and drops by the run-jump's cost on the tick it leaves the ground

#### Scenario: Too little to duck
- **WHEN** a standing fighter with less energy than a run-jump costs presses the roll button
- **THEN** it does not duck, loses no energy, and the refusal is raised

### Requirement: A hit interrupts a roll at any stage
A fighter that is ducking, rolling or recovering SHALL NOT guard. An attack that reaches it SHALL land as a hit. A stun or a death SHALL take over from the duck, the roll or the recovery, as it does from any other state.

#### Scenario: Hit while charging
- **WHEN** a ducking fighter is struck by a swing
- **THEN** the hit lands, it leaves the duck for the stun, and no energy is spent on a roll
