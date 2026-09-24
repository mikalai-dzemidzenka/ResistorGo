# energy Specification

## Purpose

Gives offence a price: a pool every swing spends and a correct guard earns back, so a fighter cannot swing forever and reading the opponent funds the next attack.

## Requirements

### Requirement: Energy pool
Every fighter SHALL have an energy pool with a capacity of 200 points, full when the fighter enters the world. Energy SHALL never exceed that capacity or fall below zero. Energy SHALL gate attacking, the running jump and the roll: walking, running, the standing jump, climbing and blocking SHALL cost nothing and SHALL be available at any energy.

#### Scenario: Full at spawn
- **WHEN** a fighter spawns, whether at the start of a match or on respawn
- **THEN** its energy is 200

#### Scenario: Restoring cannot overfill
- **WHEN** a fighter at full energy blocks an attack at the attacker's height
- **THEN** its energy is still 200

#### Scenario: An empty fighter can still defend
- **WHEN** a fighter at 0 energy is attacked
- **THEN** it can move, jump from a standstill and block normally, and blocking still resolves as it always does

#### Scenario: An empty fighter still gets around
- **WHEN** a fighter at 0 energy runs, climbs a ledge and jumps from a standstill
- **THEN** none of them is refused and none of them costs anything

#### Scenario: An empty fighter cannot roll
- **WHEN** a fighter at 0 energy presses the roll button
- **THEN** it does not duck and the refusal is raised

### Requirement: A swing costs energy
Starting a swing SHALL cost the attacker 12 energy, taken when the swing starts. Every swing SHALL cost this: the first swing of an attack and each chained swing that continues it. A swing SHALL NOT cost energy more than once however long its attack box stays over a defender, and a swing SHALL cost the same whether it lands, is blocked, or hits nothing at all.

#### Scenario: Sixteen swings from full
- **WHEN** a fighter at full energy throws swings with no time to recover between them
- **THEN** each takes its energy down by 12 — 188 after the first, 104 after the eighth, 8 after the sixteenth — and it cannot afford a seventeenth

#### Scenario: Chained swings each cost
- **WHEN** a fighter chains a second swing out of the first by holding attack
- **THEN** that second swing costs 12 energy of its own

#### Scenario: A miss costs the same
- **WHEN** a fighter swings at empty air
- **THEN** it pays the same 12 energy as a swing that lands

### Requirement: A swing that cannot be paid for does not happen
A fighter with less than 12 energy SHALL NOT start a swing. The attack input SHALL be refused: the fighter SHALL stay in the state it was already in, its animation SHALL NOT change, no attack box SHALL exist for that input, and no energy SHALL be taken. A refused attack SHALL raise a presentation event so the refusal can be shown to the player. A chained swing that cannot be paid for SHALL end the chain, leaving the fighter as it would have been had no further attack been asked for.

#### Scenario: Mashing on empty
- **WHEN** a fighter at 0 energy presses attack
- **THEN** it does not swing, stays idle, loses no energy, and the refusal event is raised

#### Scenario: Not quite enough
- **WHEN** a fighter at 8 energy presses attack
- **THEN** no swing comes out and its energy is still 8

#### Scenario: A chain runs out mid-combo
- **WHEN** a fighter holding attack reaches the end of the swing that took it to 0 energy
- **THEN** the chain ends there rather than continuing, and the fighter recovers as it does when attack is not held

#### Scenario: The refusal does not interrupt
- **WHEN** a fighter without the energy to swing presses attack while running
- **THEN** it keeps running, uninterrupted

### Requirement: A successful block restores energy
A block that holds until the swing's hit frame with a guard at the attacker's height SHALL restore 15 energy to the defender, more than the swing it buys costs. A block-all guard (guard height 3) SHALL restore no energy, however many attacks it stops. A guard that does not match and is not block-all is not a block: the attack lands as a hit and SHALL restore nothing.

#### Scenario: Reading the attacker pays
- **WHEN** a fighter at 40 energy blocks an attack with a guard at the attacker's height
- **THEN** its energy is 55

#### Scenario: Block-all does not pay
- **WHEN** a fighter at 40 energy blocks an attack while guarding with block-all
- **THEN** its energy is still 40

#### Scenario: Blocking back into the fight
- **WHEN** a fighter at 4 energy, too low to swing, blocks an attack at the attacker's height
- **THEN** its energy is 19 and it can swing immediately

#### Scenario: A hit pays nothing
- **WHEN** an attack lands on a fighter because its guard was at the wrong height
- **THEN** that fighter's energy does not change

### Requirement: Energy regenerates after a pause
When a fighter has not spent energy for 1.5 seconds, its energy SHALL refill up to capacity at one of two rates: 30 per second while attacks are still reaching it, and 60 per second once no attack has reached it for 2 seconds. An attack reaches a fighter when it lands on it or is blocked by it, block-all included. A fighter that has not been attacked since it entered the world SHALL refill at the faster rate. Spending energy SHALL restart the 1.5-second wait. Being hit or blocking SHALL restart the 2-second wait but SHALL NOT restart the 1.5-second one, and restoring energy SHALL restart neither.

#### Scenario: Recovering between exchanges
- **WHEN** a fighter that has spent 60 energy is not attacked and does not swing for 2.5 seconds
- **THEN** its energy has refilled by 60 and it is back where it was before spending

#### Scenario: Part way back
- **WHEN** a fighter that has spent 60 energy is not attacked and does not swing for 2 seconds
- **THEN** its energy has refilled by 30, half of what it spent

#### Scenario: Refilling under attack
- **WHEN** a fighter that has not swung for 1.5 seconds is attacked at least once every 2 seconds
- **THEN** its energy refills at 30 per second, half the rate of a fighter left alone

#### Scenario: A swing is soon affordable again
- **WHEN** a fighter with no energy left is not attacked and does not swing for 2 seconds
- **THEN** it can swing again

#### Scenario: A combo denies recovery
- **WHEN** a fighter swings once per second
- **THEN** its energy never regenerates between those swings and it runs out

#### Scenario: Being hit does not stop recovery
- **WHEN** a fighter that has not swung for 1.5 seconds is hit
- **THEN** its energy keeps refilling, at 30 per second

#### Scenario: Fresh fighters refill fast
- **WHEN** a fighter that has not been attacked since it spawned swings and then waits 1.5 seconds
- **THEN** its energy refills at 60 per second from the first tick of refilling

### Requirement: Energy is settled by the authority
The server's energy SHALL be the truth: every fighter's energy in a snapshot SHALL replace what a client is showing for it, and a disagreement SHALL always be resolved the server's way.

A client SHALL predict both its own fighter's spending and its regeneration, because refusing a swing decides whether that swing happens at all and the local fighter answers its own input without waiting for the server. A client that spent but did not regenerate would hold less energy than the server and so refuse swings the server allows, and the server's swing would then arrive a snapshot later as an attack the player never saw start. Everything regeneration is decided from SHALL therefore be replicated, the wait since the fighter last swung included.

The restore from a block SHALL remain the server's alone and SHALL reach a client only in a snapshot, as the block that earned it already does. A client SHALL NOT apply energy of its own to any fighter but the one it controls. Replaying commands during reconciliation SHALL spend and regenerate from the snapshot's value, so a replayed swing costs exactly once.

#### Scenario: Refusal is immediate
- **WHEN** artificial latency is 200 ms and the player clicks attack with 8 energy
- **THEN** the local fighter refuses the swing on the next simulation tick, without waiting for the server to say so

#### Scenario: A pool refilling under a held attack
- **WHEN** a client's fighter has emptied its pool with the attack button still held, and regeneration brings it back to the cost of a swing
- **THEN** the client throws that swing on the same tick the server does, and no swing appears on the client that it did not predict

#### Scenario: The block payback comes from the server
- **WHEN** a client's fighter blocks an attack at the attacker's height
- **THEN** its energy rises when the snapshot carrying the restore arrives, not before

#### Scenario: Server corrects the client
- **WHEN** a snapshot reports a different energy for a fighter than the client is showing
- **THEN** the client shows the server's value

#### Scenario: Replay does not double-charge
- **WHEN** reconciliation replays 10 commands that include one swing
- **THEN** the fighter's energy ends 12 below the snapshot's value, not 24

#### Scenario: Other fighters are displayed only
- **WHEN** another player's fighter swings on a client's screen
- **THEN** that fighter's energy changes only as snapshots report it

#### Scenario: Respawning resets it
- **WHEN** a fighter that died at 0 energy respawns
- **THEN** it is at full energy

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

### Requirement: The attack crosshair flashes when energy is won back
When the local fighter's energy rises by at least half a block's payback between one frame and the next, a short, small ring SHALL flash outward from the attack crosshair and fade. When the local fighter's energy reaches full from below, a larger, brighter, longer ring SHALL flash instead; a gain that does both SHALL show only the full flash. A new flash SHALL replace one still showing. Refilling by regeneration SHALL NOT flash on its way up. No flash SHALL start when the fighter the energy belongs to is not the one it belonged to on the previous frame — a respawn included — nor while the frame on screen is a recorded one from time manipulation. The flash SHALL be drawn with the crosshair, over the finished frame, SHALL follow the crosshair while it shows, and SHALL be shown only for the local player's own fighter. It SHALL look the same whether the energy was restored by the local simulation or by a snapshot from the server.

#### Scenario: A block pays back
- **WHEN** the local fighter, below full, blocks an attack at the attacker's height and the payback does not fill its pool
- **THEN** a small ring flashes from the attack crosshair and fades within a fraction of a second

#### Scenario: Back to full
- **WHEN** the local fighter's energy refills to full after a pause
- **THEN** a larger ring flashes from the attack crosshair, once, at the moment the pool is full

#### Scenario: A block fills the pool
- **WHEN** a block's payback brings the local fighter's energy to full
- **THEN** only the larger ring flashes

#### Scenario: Refilling is quiet
- **WHEN** the local fighter's energy is refilling and not yet full
- **THEN** no ring flashes

#### Scenario: Already full
- **WHEN** the local fighter at full energy blocks an attack at the attacker's height
- **THEN** no ring flashes

#### Scenario: Respawn is quiet
- **WHEN** the local fighter dies and respawns at full energy
- **THEN** no ring flashes

#### Scenario: Rewinding is quiet
- **WHEN** the player rewinds time across a moment the pool came back to full
- **THEN** no ring flashes while the recorded frames are shown, nor when live play resumes at the energy the fighter had

#### Scenario: Online
- **WHEN** a client's fighter blocks an attack at the attacker's height and the payback arrives in a snapshot
- **THEN** the small ring flashes when the snapshot raises the energy, exactly as it would offline

#### Scenario: Other fighters
- **WHEN** a bot or another player's fighter blocks and its energy is restored
- **THEN** nothing flashes on the local player's crosshair

### Requirement: A kill fills the killer's pool
When a fighter dies from an attack, the fighter whose attack killed it SHALL be set to full energy at that moment. This SHALL hold whether the killing attack landed or broke through a block-all, and for bots and players alike. An attack that does not kill SHALL NOT change the attacker's energy. The refill SHALL NOT restart or finish either regeneration wait. As with a block's payback, the refill SHALL be the authority's alone and SHALL reach a client only in a snapshot.

#### Scenario: A landed kill
- **WHEN** a fighter with 20 energy lands the hit that breaks its opponent's structure
- **THEN** the fighter is at full energy as its opponent falls

#### Scenario: Through a block-all
- **WHEN** a fighter's attack is blocked by an opponent guarding all heights, and paying for that block breaks the opponent's structure
- **THEN** the attacking fighter is at full energy

#### Scenario: A hit that does not kill
- **WHEN** a fighter lands a hit that leaves its opponent standing
- **THEN** the fighter's energy is unchanged by the hit

#### Scenario: Bots are rewarded too
- **WHEN** a bot kills a player
- **THEN** the bot is at full energy

#### Scenario: Online
- **WHEN** a client's fighter kills an opponent while its energy is low
- **THEN** its energy becomes full when the snapshot carrying the kill arrives, and the crosshair shows its full-pool flash
