## ADDED Requirements

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
