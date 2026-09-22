## ADDED Requirements

### Requirement: A recorded frame is the whole match
Every recorded frame SHALL hold the single-player match as it stood on that frame: which fighters were in the world, and for each its position, motion, facing, pose, state, structure, energy, stun, attack under way, buffered input, what a bot had in mind, and which fighters held the places at a fight; together with the fighters waiting to respawn and when, and the fighters dead and waiting to be taken out. While the mode is on, the frame on screen SHALL show exactly the fighters that were in the world on that frame, each as it was then.

#### Scenario: A removed fighter is back on its frame
- **WHEN** the player dies, is taken out of the world and respawns, and then scrubs back to a frame before the death
- **THEN** the fighter that died is shown where it stood on that frame, and the respawned fighter is not shown

#### Scenario: A fighter is absent before it existed
- **WHEN** the player scrubs back to a frame recorded before a fighter came into the world
- **THEN** that fighter is not shown on that frame, anywhere

#### Scenario: A body on the floor stands again on an earlier frame
- **WHEN** a bot is beaten in single player and the player scrubs back to a frame before its structure broke
- **THEN** the bot is shown standing and fighting as it was on that frame, not lying down

#### Scenario: Health bars follow the frame
- **WHEN** the player scrubs back to a frame before a fighter took a hit
- **THEN** that fighter's structure on the frame is what it was then, not what it is at the live edge

## MODIFIED Requirements

### Requirement: Leaving rewrites the future
When the player leaves the mode on a frame earlier than the live edge, the match SHALL continue from that frame exactly as recorded, as if play had stopped on it: every fighter that was in the world then SHALL be in it again with the state it had then, every fighter that was not SHALL be gone, and fighters waiting to respawn or to be taken out SHALL be waiting as they were then. Nothing that happened after that frame SHALL carry over into the resumed match — not damage, deaths, stuns, attacks, spent energy, bot decisions or places at a fight. The recorded frames that were ahead of it SHALL be discarded. A later time-manipulation session SHALL only be able to scrub back through frames recorded up to the moment it is entered, including those recorded since the match resumed. The resumed match need not repeat the discarded future: the bots may choose differently than they did.

#### Scenario: Resuming from the past discards what came after
- **WHEN** the player rewinds two seconds, leaves the mode, plays for one second and enters the mode again
- **THEN** scrubbing forward stops at the frame where they entered the second time, and the discarded frames from the first timeline are never replayed

#### Scenario: History regrows after resuming
- **WHEN** the player resumes from a rewound frame and then plays on
- **THEN** the frames played after resuming are recorded and can be rewound through in a later session

#### Scenario: Resuming before a kill undoes it
- **WHEN** the player beats a bot in single player, rewinds to a frame where it is still standing, and leaves the mode
- **THEN** the bot is alive and fighting with the structure it had on that frame, and does not fall

#### Scenario: Resuming before the player's death undoes it
- **WHEN** the player's fighter dies, respawns, and the player rewinds to a frame before the death and leaves the mode
- **THEN** the player controls the fighter that had not yet died, with the structure it had then, and no respawn is pending

#### Scenario: Resuming before a hit undoes the damage
- **WHEN** a fighter takes a hit and the player rewinds to a frame before it landed and leaves the mode
- **THEN** the fighter has the structure, energy and hit count it had on that frame, and is not stunned unless it was then

#### Scenario: A swing in progress resumes where it was
- **WHEN** the player leaves the mode on a frame in the middle of a bot's attack
- **THEN** the attack continues from that point of its animation and can land, as it would have from that frame

#### Scenario: The line at a fight is as it was
- **WHEN** the player leaves the mode on a frame where two bots held the places at the player and a third was waiting
- **THEN** the same two bots hold the places and the third is waiting, regardless of how the line had changed since
