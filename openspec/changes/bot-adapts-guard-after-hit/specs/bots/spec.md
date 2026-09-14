## Purpose

What a computer-controlled fighter does on its own: how it reacts to the fight around it, as distinct from the rules of combat, which apply to every fighter alike and belong to the combat capabilities.

## ADDED Requirements

### Requirement: A bot reacts to where it is being hit
A bot that takes a landed hit SHALL sometimes move its guard to the height the hit came in at, so that repeating one attack on it is not reliable forever. The chance SHALL be about one in three per landed hit, and SHALL be rolled independently each time: the bot SHALL NOT be made to adapt on a fixed schedule, nor guaranteed to adapt within any number of hits.

Adapting SHALL take the guard to the height the attack was made at, and nothing else about the bot's behaviour SHALL change with it. A bot that does not adapt on a given hit SHALL keep the guard it had.

This SHALL apply to a hit that flinches and to a hit that stuns. A hit that ends the fighter SHALL NOT adapt anything.

A fighter under human control SHALL NOT have its guard moved by this, at any moment and by any amount. The player's guard follows their aim and remains theirs alone, including on the tick a hit lands.

The bot's guard SHALL also continue to drift on its own, independently of being hit, so that a bot which is not being attacked still changes what it is guarding.

#### Scenario: A hit can move a bot's guard to where it landed
- **WHEN** a bot guarding low is hit by a high attack, repeatedly over many hits
- **THEN** on some of those hits its guard becomes high, and on others it stays low

#### Scenario: Adaptation is not guaranteed and not scheduled
- **WHEN** a bot takes a long run of hits at the same height
- **THEN** it neither adapts on every hit nor adapts on a fixed count of them; roughly a third of the hits move the guard and the rest leave it

#### Scenario: A stun adapts too
- **WHEN** the hit that lands is the one that stuns the bot
- **THEN** the guard may move to the attacker's height on that hit, on the same terms as any other landed hit

#### Scenario: The killing hit adapts nothing
- **WHEN** a hit breaks a bot's structure and it dies
- **THEN** its guard is not moved by that hit

#### Scenario: A player's guard is never taken
- **WHEN** a human-controlled fighter is hit at a height it is not guarding
- **THEN** its guard is exactly what the player is aiming at, on that tick and on every tick after, and any hit resolved on the same tick is judged against the guard the player chose

#### Scenario: A bot that is left alone still moves its guard
- **WHEN** a bot goes a long time without being hit
- **THEN** its guard still changes from time to time
