# clash Specification

## Purpose

Two fighters opening low swings at each other in the same instant trade blows
in a short shared exchange instead of one of them being thrown into a block.

## Requirements

### Requirement: Two low swings at once clash
Two fighters SHALL clash when all of the following hold: they are not allies; each faces the other; each is within the reach of the other's swing; and each opens a low swing within 6 ticks (100 ms) of the other. Opening a swing means starting an attack from standing or running, not a chained blow and not a change of height within a chain. A fighter already put into a low block by the other's opening low swing SHALL clash if it asks for a low swing while that block lasts and within 6 ticks of the other's swing opening, and it can pay for a swing. Swings at mid or high height, or at two different heights, SHALL NOT clash. A fighter SHALL clash with at most one other fighter at a time, and a fighter already clashing, stunned, shoved or dead SHALL NOT start a clash.

#### Scenario: Same tick
- **WHEN** two opposing fighters facing each other at arm's length both open a low swing on the same tick
- **THEN** both enter the clash, and neither blocks nor is hit

#### Scenario: Inside the window
- **WHEN** one fighter opens a low swing and the other opens one 5 ticks later, both in reach and facing each other
- **THEN** both enter the clash

#### Scenario: Caught from the block
- **WHEN** a fighter guarding low is put into a block by an opponent's opening low swing, and presses attack at low height 3 ticks after that swing opened
- **THEN** both enter the clash instead of the block playing out

#### Scenario: Outside the window
- **WHEN** the second fighter opens its low swing 10 ticks after the first
- **THEN** there is no clash, and the swings resolve as hits and blocks do

#### Scenario: Other heights do not clash
- **WHEN** two fighters open mid swings at each other on the same tick, or one low and one high
- **THEN** there is no clash

#### Scenario: A chained blow does not clash
- **WHEN** a fighter already in an attack chains its next low blow at the moment its opponent opens a low swing
- **THEN** there is no clash

#### Scenario: Back to back does not clash
- **WHEN** two fighters swing low on the same tick but one faces away from the other
- **THEN** there is no clash

### Requirement: The clash plays out on both sides together
On entering a clash, the fighter standing to the left SHALL play the left-side clash animation facing right, and the fighter to the right SHALL play the right-side clash animation facing left. Both animations SHALL last the same time and start on the same tick, so they end together. The two fighters SHALL be placed facing each other at the spacing the animations are drawn for, not moving through walls. Neither fighter SHALL be steered, turn, jump, attack or block while clashing, and input given during a clash SHALL NOT be carried out after it. When its animation ends, each fighter SHALL return to standing, free to act.

#### Scenario: Sides follow position
- **WHEN** a clash starts between a fighter on the left and one on the right
- **THEN** the left one plays the lunging punch facing right and the right one plays the knee facing left

#### Scenario: Ending together
- **WHEN** a clash has started
- **THEN** both fighters leave it on the same tick and are standing and able to act

#### Scenario: No steering while clashing
- **WHEN** the player holds a movement key and clicks attack during a clash
- **THEN** the fighter neither moves nor attacks during the clash, and does not attack when it ends

### Requirement: What a clash costs
Entering a clash SHALL cost each fighter 0.5 structure and one swing's energy. A fighter whose opening swing already paid for a swing SHALL NOT pay again; a fighter that clashed from a block SHALL pay one swing's energy on entering. A clash SHALL NOT flash either fighter, pause the match for a hit, knock either fighter back, flinch or stun either, count toward either one's hits before a stun, or add to either one's run of hits. Structure spent in a clash SHALL delay regeneration as any other loss does. A fighter whose structure a clash empties SHALL die as any fighter whose structure breaks does, and the other fighter SHALL be credited with the kill; the clash SHALL then not play for either fighter, and a surviving fighter SHALL return to standing.

#### Scenario: Both pay
- **WHEN** two fighters at full structure and full energy clash
- **THEN** each has 11.5 structure and 12 less energy than before its swing, and neither flashes

#### Scenario: Not a hit
- **WHEN** a fighter that has taken two hits since its last stun clashes
- **THEN** it is not stunned, and the next hit it takes is still its third

#### Scenario: A clash can kill
- **WHEN** a fighter with 0.5 structure left clashes
- **THEN** it dies, the other fighter is credited with the kill, and no clash animation plays

### Requirement: Clashing fighters are out of reach of everyone else
While a fighter is clashing, no other fighter's swing SHALL hit it, be blocked by it, or put it into a block, and no fighter SHALL shove it or be shoved by it. A swing that reaches a clashing fighter SHALL pass through it as it passes through an ally.

#### Scenario: A third fighter swings at a clash
- **WHEN** a bot swings at a fighter that is clashing with another bot
- **THEN** the swing has no effect on either clashing fighter

#### Scenario: Protection ends with the clash
- **WHEN** a clash ends and a third fighter's swing reaches one of the two afterwards
- **THEN** the swing is resolved as a hit or a block as usual

### Requirement: Clashes are decided by the authority
Whether two swings clash, and what the clash costs, SHALL be decided only by the authoritative simulation, as hits and blocks are. A client SHALL NOT start a clash of its own; it SHALL show a clash, for its own fighter as for any other, as the server reports it.

#### Scenario: Online clash
- **WHEN** a player and a bot open low swings at each other within the window in a networked match
- **THEN** the server puts both into the clash, and the player's client shows its fighter clashing once the snapshot arrives, with the server's structure and energy
