# firearms Specification

## Purpose
Lets a fighter draw a gun, aim it with the mouse and fire shots that stop at walls and land hits on fighters. For now the pistol is the only gun.

## Requirements

### Requirement: The pistol is drawn and holstered with a key

A fighter SHALL start every life with the pistol holstered. Pressing Q while it stands, walks or runs on the ground SHALL draw the pistol, and pressing Q again under the same conditions SHALL holster it. A press in any other state SHALL do nothing and SHALL NOT be remembered for later. The pistol SHALL stay drawn through a jump, a fall, a climb, a roll, a hit or a stun, and SHALL be drawn again when the fighter next stands. Bots SHALL never draw a gun. Q SHALL keep its editor meaning while the editor is open.

#### Scenario: Drawing the pistol
- **WHEN** a standing fighter with the pistol holstered presses Q
- **THEN** it holds the pistol from that tick

#### Scenario: Holstering the pistol
- **WHEN** a standing fighter holding the pistol presses Q
- **THEN** it puts the pistol away and stands with its usual guard

#### Scenario: No draw in the air
- **WHEN** a jumping fighter presses Q
- **THEN** nothing changes, then or when it lands

#### Scenario: The gun survives a jump
- **WHEN** a fighter holding the pistol jumps and lands
- **THEN** it plays the usual jump, and holds the pistol again once it stands

#### Scenario: A new life starts unarmed
- **WHEN** a fighter that died holding the pistol respawns
- **THEN** it has the pistol holstered

### Requirement: The pistol is aimed at the mouse

While a fighter holds the pistol, it SHALL face the side the mouse is on, and its aim SHALL be the angle from the fighter to the mouse, measured from horizontal on the side it faces. The aim SHALL be clamped to the range the pistol sheet covers, from its highest direction to its lowest. The torso SHALL be drawn with whichever of the sheet's five directions, from up to down, is nearest the aim.

#### Scenario: Aiming level
- **WHEN** the mouse is level with the fighter's shoulder, to its right
- **THEN** the fighter faces right and is drawn with the middle, horizontal direction

#### Scenario: Aiming high
- **WHEN** the mouse is well above the fighter, ahead of it
- **THEN** the fighter is drawn with the highest direction, and the aim goes no higher than that direction

#### Scenario: Aiming behind
- **WHEN** the mouse moves from ahead of the fighter to behind it
- **THEN** the fighter turns to face the mouse and keeps aiming at it

### Requirement: A fighter holding the pistol is drawn in two layers

While a fighter holds the pistol and stands or walks, it SHALL be drawn as the gun legs layer with the pistol torso on top, at the aim direction's frame, facing the way the fighter faces. In every other state it SHALL be drawn with that state's usual animation and no pistol. The rim light SHALL trace the torso layer as well as the legs.

#### Scenario: Standing with the pistol
- **WHEN** a fighter holding the pistol stands still
- **THEN** it is drawn as the idle gun legs with the pistol torso over them

#### Scenario: Walking with the pistol
- **WHEN** a fighter holding the pistol walks
- **THEN** it moves with the idle gun legs frame, and the torso follows the aim

#### Scenario: Hit while holding the pistol
- **WHEN** a fighter holding the pistol is stunned
- **THEN** it plays the usual hit animation without the pistol

### Requirement: A fighter holding the pistol walks and does not guard

While a fighter holds the pistol, it SHALL be able to walk but SHALL NOT run: its speed SHALL be held to walking pace. It SHALL NOT swing, block or guard block-all. A swing against it SHALL land at any height. Jumping, rolling and climbing down SHALL work as they do unarmed.

#### Scenario: No running with the gun
- **WHEN** a fighter holding the pistol holds a movement key
- **THEN** it walks at walking pace and never enters the run

#### Scenario: Drawing mid-run
- **WHEN** a running fighter presses Q
- **THEN** it holds the pistol and slows to walking pace

#### Scenario: No guard with the gun
- **WHEN** an enemy swings high at a fighter holding the pistol while that fighter's mouse is high
- **THEN** the swing lands as a hit and is not blocked

### Requirement: LMB fires the pistol

While a fighter holds the pistol and stands or walks, pressing LMB SHALL fire one shot along its aim, in place of a swing. After a shot the pistol SHALL NOT fire again for a quarter of a second, and a press within that time SHALL be ignored, not stored. A shot SHALL cost no energy. A press in any other state SHALL do nothing.

#### Scenario: Firing
- **WHEN** a standing fighter holding the pistol clicks LMB
- **THEN** one shot is fired on that tick along the aim, and no swing starts

#### Scenario: Cooldown
- **WHEN** a fighter fires and clicks LMB again 0.1 seconds later
- **THEN** no second shot is fired, and none is fired later on account of that click

#### Scenario: Energy is not spent
- **WHEN** a fighter holding the pistol with an empty energy pool fires
- **THEN** the shot is fired and its energy is unchanged

### Requirement: A shot stops at the first wall or fighter

A shot SHALL travel in a straight line from the pistol's muzzle along the aim, up to 240 pixels. It SHALL stop at the nearest of: the first solid level rectangle it crosses (drop-through platforms do not stop it), or the collision box of the first living fighter that is not the shooter's ally. Dead fighters and allies SHALL NOT stop it. A shot that meets nothing SHALL end at its full range.

#### Scenario: Stopped by a wall
- **WHEN** a fighter fires at a wall with an enemy standing behind it
- **THEN** the shot ends at the wall's face, and the enemy is not hit

#### Scenario: Stopped by the nearer fighter
- **WHEN** two enemies stand one behind the other along the aim
- **THEN** the shot ends at the nearer one's collision box, and only that one is hit

#### Scenario: Passing an ally
- **WHEN** an ally stands between the shooter and an enemy along the aim
- **THEN** the shot passes the ally and stops at the enemy

#### Scenario: Through a drop-through platform
- **WHEN** a fighter fires down through a drop-through platform at an enemy below it
- **THEN** the shot passes the platform and stops at the enemy

### Requirement: A shot lands a hit

A fighter a shot stops at SHALL take a hit from the shooter, the same hit a landed swing gives: the same structure cost and hit run, a flinch, a stun on every third hit, and death with the kill credited to the shooter when its structure runs out. Guard SHALL NOT stop a shot, whatever its height and including block-all. A fighter that is blocking SHALL leave the block and take the hit. A fighter already stunned or clashing SHALL stop the shot and take no hit from it. Only the authoritative simulation SHALL decide whether a shot hits. Unlike a landed swing, a landed shot SHALL NOT freeze the shooter in hit-stop; only the target is frozen.

#### Scenario: Shot costs structure
- **WHEN** a shot stops at an enemy with full structure
- **THEN** the enemy loses the structure a first hit costs, and flinches

#### Scenario: Block-all does not stop a shot
- **WHEN** a shot stops at an enemy guarding block-all
- **THEN** it lands as a hit

#### Scenario: Kill by a shot
- **WHEN** a shot empties an enemy's structure
- **THEN** the enemy dies, and the shooter is credited with the kill

#### Scenario: The shooter is not frozen
- **WHEN** a shot lands on an enemy
- **THEN** the enemy is frozen in hit-stop and the shooter goes on moving on the next tick

#### Scenario: Stunned target
- **WHEN** a shot stops at an enemy that is already stunned
- **THEN** the shot ends there and the enemy's structure is unchanged

### Requirement: Every shot is shown with a tracer and a muzzle flash

Every shot SHALL be shown, to every player and spectator watching, as a thin bright tracer from the muzzle to where the shot stopped, lasting about 0.06 seconds, and a muzzle flash: a small warm light at the muzzle, lasting about 0.05 seconds, that lights the scene and casts shadows like a level light. The tracer SHALL be drawn over the lit scene so the ambient dark does not dim it. A shot SHALL be shown once, however many times the local simulation replays the tick it was fired on.

#### Scenario: Own shot
- **WHEN** the local player fires
- **THEN** a tracer and a flash appear on the tick the shot is fired, without waiting for the server

#### Scenario: Another player's shot
- **WHEN** another player fires
- **THEN** a tracer and a flash appear at that player's muzzle once the snapshot carrying the shot arrives

#### Scenario: Flash in a dark area
- **WHEN** a fighter fires in an unlit part of the level
- **THEN** the area around the muzzle is briefly lit, and then dark again

#### Scenario: Correction does not repeat a shot
- **WHEN** the client corrects its fighter and replays ticks that include a shot
- **THEN** that shot's tracer and flash are not shown a second time

### Requirement: A shot leaves smoke, and dust where it hits a wall

Every shot, wherever it is shown, SHALL leave a few puffs of grey smoke at the muzzle. The puffs SHALL leave along the shot's direction, slow down, drift upward, grow, and thin out pixel by pixel until they are gone, within about a second. A shot that stops at a wall SHALL also raise a few smaller puffs of dust at the point it hit, drifting back out of the wall and up, and gone sooner than the smoke. A shot that stops at a fighter or runs out of range SHALL raise no dust. Smoke and dust SHALL be drawn inside the lit scene, so the muzzle flash lights them and the level's dark darkens them. Like the tracer, they SHALL be shown once per shot and SHALL have no effect on the simulation.

#### Scenario: Smoke after a shot
- **WHEN** a fighter fires
- **THEN** smoke puffs appear at the muzzle, drift up and away, and are gone within about a second

#### Scenario: Dust on a wall
- **WHEN** a shot stops at a wall
- **THEN** dust puffs appear where it hit, and drift out of the wall

#### Scenario: No dust on a fighter
- **WHEN** a shot stops at a fighter
- **THEN** there is smoke at the muzzle but no dust where it hit

#### Scenario: Smoke in the dark
- **WHEN** a fighter fires in an unlit part of the level
- **THEN** the smoke is lit by the flash for a moment, then fades into the dark as it drifts

### Requirement: The pistol kicks when fired

For about 0.1 seconds after a shot, the torso SHALL be drawn one direction higher than the aim (or, when the aim is already at its highest, shifted a pixel back), and then return to the aim. Every machine SHALL show the kick from replicated state, without a message of its own. The local player's own shot SHALL also kick the camera a little against the aim, easing back within about a tenth of a second.

#### Scenario: Kick after a shot
- **WHEN** a fighter aiming level fires
- **THEN** its torso shows the direction above level for about 0.1 seconds, then level again

#### Scenario: Another player's kick
- **WHEN** another player fires
- **THEN** its torso kicks on every client as it does on the shooter's screen

#### Scenario: Camera kick
- **WHEN** the local player fires
- **THEN** the view jolts slightly away from the aim and settles back

### Requirement: The muzzle shows a flash of fire

Every shot SHALL draw a small bright star of pixels at the muzzle, longest along the shot, for about 0.05 seconds, over the lit scene so the dark does not dim it.

#### Scenario: Fire at the muzzle
- **WHEN** a fighter fires
- **THEN** a bright star of pixels appears at the muzzle for a few frames

### Requirement: A shot ejects a casing

Every shot SHALL throw a small brass casing out of the pistol, back over the shooter's shoulder and up. It SHALL fall, bounce off level solids with most of its speed lost, come to rest, and disappear a few seconds later. Casings SHALL have no effect on the simulation.

#### Scenario: Casing falls and rests
- **WHEN** a fighter standing on the floor fires
- **THEN** a casing flies back and up, bounces on the floor, comes to rest on it, and is gone a few seconds later

### Requirement: A shot that hits a wall sparks and leaves a hole

A shot that stops at a wall SHALL throw a few bright sparks back out of the wall, gone within about 0.15 seconds, and SHALL leave a small dark bullet hole at the point it hit, which stays for a while and then disappears. Only a limited number of holes SHALL be kept, the oldest going first. A shot that stops at a fighter or runs out of range SHALL do neither. Holes, casings and every other shot effect SHALL be cleared when a match starts over.

#### Scenario: Sparks and a hole
- **WHEN** a shot stops at a wall
- **THEN** sparks fly back out of the wall, and a dark hole stays where it hit

#### Scenario: No hole in a fighter
- **WHEN** a shot stops at a fighter
- **THEN** no sparks and no hole appear

### Requirement: Every shot is heard

Every shot SHALL play a gunshot sound, picked at random from the pistol's sounds, at the sound volume setting, once per shot on every machine that shows it. Shots fired close together SHALL each be heard rather than cut one another short. With no audio device the game SHALL run silently.

#### Scenario: A shot is heard
- **WHEN** any fighter fires
- **THEN** one gunshot sound plays

#### Scenario: Rapid shots
- **WHEN** a fighter fires again while its last shot's sound is still playing
- **THEN** the new shot is heard without stopping the one before
