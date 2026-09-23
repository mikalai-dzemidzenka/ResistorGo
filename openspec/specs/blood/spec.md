# blood Specification

## Purpose

Makes the fight leave a mark: a landed blow sprays blood away from it, and the
blood stays where it lands for the rest of the match.

## Requirements

### Requirement: A landed blow spills blood
When a fighter's structure drops because a hit landed on it, blood SHALL spray from its body, thrown mostly away from the fighter that hit it. A hit that flinches SHALL spray a little, a hit that stuns SHALL spray more, and a hit that kills SHALL spray the most. Structure lost to a guarded block, to block-all, to a clash or to anything but a landed hit SHALL spill nothing, and neither SHALL a shove. Structure rising SHALL spill nothing.

#### Scenario: A flinch spills a little
- **WHEN** a hit lands on a fighter and does not stun it
- **THEN** a few droplets fly from it, away from the attacker

#### Scenario: More for a stun, most for a kill
- **WHEN** one hit flinches a fighter, another stuns one, and a third kills one
- **THEN** the stun sprays more droplets than the flinch, and the kill more than the stun

#### Scenario: Blocked blows are clean
- **WHEN** a fighter blocks a swing, or pays structure for block-all, or clashes
- **THEN** no blood is spilled

### Requirement: Blood lands and stays
Each droplet SHALL fly under gravity from where it was sprayed, for a short distance of its own chosen at random when it is sprayed, and then splash onto the wall behind the fight: where it stops over a painted pixel of a level tile, it SHALL become a streak on that tile - a clump at its head with a short smear along the way it was flying - and the clump SHALL cover only the tile's painted pixels. Where it stops over empty background or a transparent part of a tile, it SHALL fly on. A droplet that meets the level's collision - a floor, a wall or a ceiling - before that SHALL stop there and become a stain at that point: a clump smeared along that surface, covering only the collision and the pixels touching it, never hanging in the air beside it. Every stain and streak SHALL stay in place, unchanged, for the rest of the match. A droplet SHALL NOT hit or be stopped by fighters. A droplet that leaves the level without landing SHALL disappear. The number of stains and streaks SHALL be capped together; once the cap is reached, each new one SHALL replace the oldest. A new match, and restarting the level, SHALL start with no blood.

#### Scenario: Splashed on the wall behind
- **WHEN** a fighter standing in front of a backdrop wall tile is hit
- **THEN** most of the blood splashes onto that tile behind it as streaks pointing away from the blow, and stays there

#### Scenario: Not in the void
- **WHEN** a fighter standing against empty background with no tile behind it is hit
- **THEN** no blood hangs in the empty air; it falls on and stains the floor

#### Scenario: Only on the painted part
- **WHEN** a droplet stops over the transparent part of a tile
- **THEN** it does not stick there, and flies on

#### Scenario: A stain where it fell
- **WHEN** a fighter is hit standing on the floor
- **THEN** the droplets arc away and land as stains on the floor near it, and are still there a minute later

#### Scenario: Blood on the wall
- **WHEN** a fighter is hit with a wall close behind it, on the side away from the attacker
- **THEN** some of the droplets stain the wall

#### Scenario: A long fight
- **WHEN** a match goes on until more stains have landed than the cap allows
- **THEN** the newest stains are all shown and the oldest are gone, and the frame rate is unchanged

#### Scenario: A fresh level
- **WHEN** the player restarts the level after a bloody fight
- **THEN** the floor is clean

### Requirement: Blood is drawn in the level's light
Blood SHALL be drawn in the level's reds, as part of the lit scene, so that it is darker in shadow and brighter in light like the tiles around it. Stains and streaks SHALL be drawn over the level and under the fighters; droplets in flight SHALL be drawn over the fighters. A droplet SHALL be a clump of one to three level pixels in one piece, and it SHALL splash out into a bigger clump where it lands. A clump SHALL be no wider or taller than five pixels, and darker at its rim than at its middle; a streak's tail SHALL be one to four pixels in a line behind its clump, darker than its head.

#### Scenario: In the dark
- **WHEN** blood lands in an unlit corner
- **THEN** it is a dark red there, not the bright red of the same blood under a light

### Requirement: Blood follows time manipulation
While time manipulation shows a recorded frame, the stains shown SHALL be exactly those that had landed by that frame, and droplets SHALL neither fly nor be sprayed. Leaving the mode on a frame SHALL keep exactly those stains, so the match resumes with the blood it had on that frame. While the match is held with nothing moving, droplets in flight SHALL hold still.

#### Scenario: Rewinding takes the blood back
- **WHEN** the player rewinds to before a kill
- **THEN** the kill's stains are gone, and the stains from earlier in the fight are still there

#### Scenario: Scrubbing sprays nothing
- **WHEN** the player scrubs forward and back across a hit
- **THEN** no droplets appear while scrubbing

#### Scenario: Resuming keeps the frame's blood
- **WHEN** the player leaves the mode on a frame before a kill and the match goes on
- **THEN** the kill's stains stay gone, and new blood is spilled only by the hits that happen from there

### Requirement: Every machine makes its own blood
Blood SHALL be presentation only: it SHALL NOT change the simulation, be sent over the network, or be decided by the server. A client SHALL spill blood when the structure it is shown for a fighter drops, for its own fighter and every other, and a spectator likewise. Where the droplets land MAY differ from one machine to another. A dedicated server with no screen SHALL keep no blood.

#### Scenario: Online
- **WHEN** a player is hit by a bot in a networked match
- **THEN** both the player and a spectator see blood spill from the player, away from the bot

#### Scenario: The match is unchanged
- **WHEN** the same fight is simulated with blood and without it
- **THEN** every fighter's state is identical tick for tick
