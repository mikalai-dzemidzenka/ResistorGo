## MODIFIED Requirements

### Requirement: A landed hit costs structure and flinches
A hit that lands SHALL cost the defender structure and SHALL produce a flinch: the hit flash, the hit-stop freeze that both fighters already share, and a small knockback. A flinch SHALL NOT play an animation of its own: the defender SHALL keep whatever animation it was already playing, and SHALL be shown to have been hit by the flash alone. A flinch SHALL NOT lock the defender out of acting — it can move, attack, block and jump during one — and SHALL NOT interrupt an action the defender had already started.

The flash SHALL hold for about 0.2 seconds and then stop on its own. It SHALL also stop early with whatever replaces the flinch: a fighter that swings, blocks, is stunned or dies while flashing SHALL stop flashing at that moment, so no fighter is ever left lit.

What a hit costs SHALL depend on how many hits have already landed on that fighter in the same run. The first hit of a run SHALL cost 1 structure; each hit after it SHALL cost half a point more than the one before — 1, 1.5, 2, 2.5, 3 — and the cost SHALL NOT climb past 3 however long the run goes on. The run SHALL be per fighter: it counts what has landed on that fighter, whoever landed it.

The knockback SHALL move the fighter, not merely be given to it: after a flinch the fighter SHALL be somewhere it was not when the hit landed. It SHALL remain clearly smaller than the throw a stun gives, so the two read as different events.

The hit-stop freeze SHALL hold input, never discard it. An attack or a jump pressed while a fighter is frozen SHALL be held for that fighter and offered to its first tick after the freeze ends, alongside anything pressed on that tick, and SHALL then take effect only as far as the fighter's state at that moment allows. This holds for both fighters the freeze covers: the defender's counter and the attacker's next swing are equally kept. A freeze is therefore unlike a paused mode, which is not holding input at all — while a paused mode owns the mouse and keyboard the press belongs to that mode, and the simulation's latch is cleared because it was never the intended consumer.

Input held across a freeze SHALL NOT outlive it: whatever the first tick after the freeze does with the press, held or acted on, the fighter SHALL NOT be holding it afterwards. A held button SHALL NOT produce more than the one action it was pressed for, however many frozen ticks it spans. A fighter that dies SHALL discard anything it was holding, so nothing crosses into the fighter that replaces it.

#### Scenario: First hit does not stun
- **WHEN** a fighter at full structure takes one hit
- **THEN** its structure is 11, it flashes and is knocked back slightly, and it can act on its first tick after the freeze

#### Scenario: The small knockback actually moves the fighter
- **WHEN** a fighter is flinched and then left alone until the knockback has run out
- **THEN** its position differs from where it was when the hit landed

#### Scenario: A flinch stays smaller than a stun
- **WHEN** one fighter is flinched and another stunned by attackers facing the same way, and both are left alone until their knockbacks have run out
- **THEN** the flinched one has travelled a visibly shorter distance than the stunned one

#### Scenario: A run costs more as it goes
- **WHEN** a fighter at full structure takes five hits in a row with nothing to break the run
- **THEN** they cost 1, 1.5, 2, 2.5 and 3, leaving it at 2 structure

#### Scenario: The cost stops climbing
- **WHEN** a run reaches a sixth and a seventh hit
- **THEN** each costs 3, the same as the fifth

#### Scenario: Six hits end a fight
- **WHEN** a fighter at full structure takes six hits in a row
- **THEN** its structure reaches zero and it dies

#### Scenario: The stun does not relieve it
- **WHEN** a fighter takes four hits in a row, the third of which stuns it
- **THEN** the fourth costs 2.5, as the fourth hit of a run does, rather than starting the run again

#### Scenario: Time ends a run
- **WHEN** a fighter is left alone long enough for its structure to start refilling, and is then hit again
- **THEN** that hit costs 1, as the first hit of a run does

#### Scenario: Flinch does not interrupt a swing
- **WHEN** a fighter is hit while its own attack animation is playing
- **THEN** that attack continues and can still land

#### Scenario: A flinch leaves the animation alone
- **WHEN** a running fighter takes a hit that does not stun it
- **THEN** it goes on running, with the running animation uninterrupted, and flashes

#### Scenario: The flash runs out
- **WHEN** a fighter has been flinched and half a second passes with nothing else happening to it
- **THEN** it is no longer flashing

#### Scenario: Acting ends the flash
- **WHEN** a flinched fighter swings while it is still flashing
- **THEN** the flash stops as the swing starts

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

### Requirement: Every third landed hit stuns
The third hit to land on a fighter SHALL stun it: the defender enters the hit-stun state, is locked out until the stun animation finishes, and is thrown back noticeably further than a flinch throws it. The stun SHALL be the only hit that plays the hit animation, so the animation tells the player their fighter has lost control rather than merely been hit; the flash SHALL cover the opening of that animation, not the whole of it. The count SHALL be per fighter and SHALL reset when the stun lands, when the fighter dies, and when its structure begins regenerating. A block SHALL NOT advance it.

#### Scenario: Third hit of a run
- **WHEN** a fighter takes three hits without its structure regenerating in between
- **THEN** the first two flinch it and the third stuns it, locking it out and throwing it back further

#### Scenario: Only the stun has an animation
- **WHEN** a fighter takes three hits in a row
- **THEN** neither of the first two changes the animation it was playing, and the third plays the hit animation

#### Scenario: The count restarts after a stun
- **WHEN** a fighter has just been stunned and takes another hit
- **THEN** that hit flinches it, and the stun comes two hits later

#### Scenario: Blocks do not count
- **WHEN** a fighter takes a hit, blocks two attacks, then takes another hit
- **THEN** neither hit stuns it
