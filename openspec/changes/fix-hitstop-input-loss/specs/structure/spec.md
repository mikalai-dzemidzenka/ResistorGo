## MODIFIED Requirements

### Requirement: A landed hit costs structure and flinches
A hit that lands SHALL cost the defender 1 structure and SHALL produce a flinch: the hit flash, the hit-stop freeze that both fighters already share, a small knockback and the hit animation. A flinch SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one — and SHALL NOT interrupt an action the defender had already started.

The hit-stop freeze SHALL delay input, never discard it. An attack or a jump pressed while a fighter is frozen SHALL take effect on that fighter's first tick after the freeze ends, subject to whatever its state then allows. This holds for both fighters the freeze covers: the defender's counter and the attacker's next swing are equally kept. A freeze is therefore unlike a paused mode, which discards the input given while it holds the match, because a click during a freeze was aimed at the fight and a click during a pause was not.

A held button SHALL NOT produce more than the one action it was pressed for, however many frozen ticks it spans.

#### Scenario: First hit does not stun
- **WHEN** a fighter at full structure takes one hit
- **THEN** its structure is 11, it flashes and is knocked back slightly, and it can act on its first tick after the freeze

#### Scenario: Flinch does not interrupt a swing
- **WHEN** a fighter is hit while its own attack animation is playing
- **THEN** that attack continues and can still land

#### Scenario: A counter pressed during the freeze still comes out
- **WHEN** a fighter is hit and the player presses attack during the hit-stop freeze
- **THEN** the fighter attacks on its first tick after the freeze, and pressing attack a moment later has the same result

#### Scenario: The attacker's next swing is kept too
- **WHEN** a swing lands and the player presses attack during the freeze it causes
- **THEN** that press is queued as the chain's next swing, exactly as a press later in the same swing would be

#### Scenario: One press is one action
- **WHEN** the player presses attack once during a freeze that lasts several ticks
- **THEN** exactly one attack results

#### Scenario: A stun still locks the defender out
- **WHEN** a fighter is stunned and the player presses attack during the hit-stop freeze
- **THEN** the fighter does not attack, and the press does not survive the stun to fire when it ends
