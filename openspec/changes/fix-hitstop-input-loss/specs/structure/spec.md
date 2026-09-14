## MODIFIED Requirements

### Requirement: A landed hit costs structure and flinches
A hit that lands SHALL cost the defender 1 structure and SHALL produce a flinch: the hit flash, the hit-stop freeze that both fighters already share, a small knockback and the hit animation. A flinch SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one — and SHALL NOT interrupt an action the defender had already started.

The hit-stop freeze SHALL hold input, never discard it. An attack or a jump pressed while a fighter is frozen SHALL be held for that fighter and offered to its first tick after the freeze ends, alongside anything pressed on that tick, and SHALL then take effect only as far as the fighter's state at that moment allows. This holds for both fighters the freeze covers: the defender's counter and the attacker's next swing are equally kept. A freeze is therefore unlike a paused mode, which is not holding input at all — while a paused mode owns the mouse and keyboard the press belongs to that mode, and the simulation's latch is cleared because it was never the intended consumer.

Input held across a freeze SHALL NOT outlive it: whatever the first tick after the freeze does with the press, held or acted on, the fighter SHALL NOT be holding it afterwards. A held button SHALL NOT produce more than the one action it was pressed for, however many frozen ticks it spans. A fighter that dies SHALL discard anything it was holding, so nothing crosses into the fighter that replaces it.

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

#### Scenario: A jump pressed during the freeze is kept
- **WHEN** a fighter is hit and the player presses jump during the hit-stop freeze
- **THEN** the fighter leaves the ground on its first tick after the freeze, as it would have done had the press landed on that tick

#### Scenario: One press is one action
- **WHEN** the player presses attack once during a freeze that lasts several ticks
- **THEN** exactly one attack results

#### Scenario: A stun still locks the defender out
- **WHEN** a fighter is stunned and the player presses attack during the hit-stop freeze
- **THEN** the fighter does not attack, and the press does not survive the stun to fire when it ends

#### Scenario: A press cannot be afforded away while it waits
- **WHEN** a fighter with exactly enough energy for one swing is hit and the player presses attack during the freeze
- **THEN** the swing comes out on the first tick after the freeze and is paid for, because nothing spent or refilled the pool while the fighter was frozen

#### Scenario: Death discards a held press
- **WHEN** the hit that breaks a fighter's structure is the one during whose freeze the player presses attack
- **THEN** the fighter dies holding nothing, and the press does not act for it or for the fighter the player is given next
