## Purpose

What a computer-controlled fighter does on its own: how it reacts to the fight around it, as distinct from the rules of combat, which apply to every fighter alike and belong to the combat capabilities.

## ADDED Requirements

### Requirement: A bot acts only on fighters it can see
A bot SHALL choose the fighter it reacts to from those it can see, and only from those. Seeing another fighter SHALL require all three of:

- **Range** — the other fighter is within the bot's sight range.
- **Direction** — the other fighter lies within a fixed angle either side of the direction the bot is facing. A fighter behind the bot SHALL NOT be seen, and neither SHALL one far enough above or below that it falls outside that angle.
- **A clear line** — no solid part of the level lies on the straight line between the two fighters. A fighter on the far side of a wall, a floor or a ceiling SHALL NOT be seen, however close it is and however squarely the bot faces it.

When more than one fighter is seen, the bot SHALL react to the nearest of them. Sight SHALL be judged afresh every tick, so a fighter stepping behind cover is lost the moment the cover is between them, and one stepping out is seen again the moment it is not.

Sight SHALL be a bot's alone. A fighter under human control SHALL never have what it may do limited by any of this, and SHALL go on being controlled entirely by the player's input.

#### Scenario: A fighter in front, close and in the open is seen
- **WHEN** a fighter stands within sight range, ahead of the bot and level with it, with nothing between them
- **THEN** the bot reacts to it — it approaches and swings as it always has

#### Scenario: A fighter behind the bot is not seen
- **WHEN** a fighter stands within sight range and in the open, but on the side the bot has its back to
- **THEN** the bot does not react to it: it neither approaches it nor turns toward it

#### Scenario: A fighter beyond the range is not seen
- **WHEN** a fighter stands directly ahead of the bot, in the open, but further away than the sight range
- **THEN** the bot does not react to it, and does react to it once it comes inside the range

#### Scenario: A wall breaks the line
- **WHEN** a fighter stands ahead of the bot and within range, with a solid part of the level between them
- **THEN** the bot does not react to it, and does react to it again once the fighter steps out from behind that solid

#### Scenario: A fighter far enough above is outside the cone
- **WHEN** a fighter stands within range but steeply above or below the bot, outside the angle either side of the way the bot faces
- **THEN** the bot does not react to it

#### Scenario: The nearest of several
- **WHEN** a bot can see two fighters at once
- **THEN** it reacts to the nearer one

#### Scenario: A player is never blinded
- **WHEN** a human-controlled fighter has an opponent behind it, out of range or behind a wall
- **THEN** nothing about what that fighter may do changes: it moves, turns, jumps and attacks exactly as the player asks

### Requirement: A bot that sees no one does not fight
A bot with no fighter in sight SHALL NOT approach anyone, SHALL NOT swing, and SHALL NOT turn to face anyone. It SHALL remain a fighter at rest: standing where it is once it has come to a stop, and still drifting its guard height from time to time as it does when nothing is happening to it.

Losing sight SHALL NOT interrupt anything already committed. A swing under way SHALL finish, a jump already left the ground SHALL land, a climb SHALL complete, and a bot being hit SHALL go on reacting to the hit exactly as it would with the attacker in plain view.

#### Scenario: Alone in the level
- **WHEN** a bot has no fighter within sight
- **THEN** it does not walk toward anyone or swing at anything, and its guard still changes from time to time

#### Scenario: A swing already thrown finishes
- **WHEN** a bot's target disappears from sight on the tick after the bot started a swing
- **THEN** the swing plays out and resolves as any swing does

#### Scenario: Still reacting to being hit
- **WHEN** a bot is hit by an attacker it cannot see
- **THEN** it takes the hit, and its structure, stun and guard react exactly as they would to a hit from a visible attacker

### Requirement: An attack tells a bot where its attacker is
An attack that reaches a bot SHALL tell it where the attacker was standing, whether the bot could see the attacker or not. A bot SHALL take that position as the place it last saw a fighter at, so that a blow out of nowhere turns it and sends it to look — and where the attacker is still standing there, the bot SHALL see them on arriving and fight back on the ordinary terms.

This SHALL apply to an attack that lands and to one the bot blocks: a blow it guarded says as much about where someone is standing as one that hurt.

A bot that can already see a fighter SHALL be unaffected by this: it is reacting to what is in front of it, and what it can see SHALL continue to outrank anything it merely remembers.

Being hit SHALL NOT make an attacker visible. The bot SHALL learn a position and nothing else — it SHALL NOT swing at, track, or turn with a fighter it cannot see, and an attacker who moves away after striking SHALL leave the bot searching the place the blow came from rather than following them.

A fighter under human control SHALL NOT be steered by any of this.

#### Scenario: A blow from behind turns the bot round
- **WHEN** a bot is attacked by a fighter standing behind it, outside its cone
- **THEN** it turns toward where that attack came from and goes to look, rather than standing on unaware

#### Scenario: And then it fights
- **WHEN** the attacker is still standing there as the bot comes round
- **THEN** the bot sees it and fights back exactly as it would a fighter it had spotted for itself

#### Scenario: A blocked attack tells it the same
- **WHEN** a bot blocks an attack from a fighter it cannot see
- **THEN** it learns where that attacker was standing, on the same terms as an attack that landed

#### Scenario: A hit does not make the attacker visible
- **WHEN** a bot is struck by a fighter that then moves out of the bot's way
- **THEN** the bot goes to the place the blow came from rather than following the attacker, and gives the search up there as it would any other

#### Scenario: What it can see comes first
- **WHEN** a bot already has a fighter in view and is struck from behind by a second one
- **THEN** it goes on reacting to the fighter it can see

### Requirement: A bot searches where it last saw someone
A bot that loses sight of the fighter it was reacting to SHALL remember the position that fighter was last seen at and go there. It SHALL move toward that position by the same rules it uses to move anywhere, stopping where the level stops it. On arriving — or once it has spent its patience trying to arrive, whichever comes first — the bot SHALL wait there briefly and then give up, returning to the resting behaviour of a bot that sees no one.

A search SHALL end immediately if the bot sees any fighter while it is under way; the bot SHALL react to what it now sees rather than to where something used to be. A bot SHALL search for one remembered position at a time: losing sight again replaces the remembered position rather than queueing another search.

A searching bot SHALL NOT swing. It is going somewhere, not fighting.

#### Scenario: Walking to the last known spot
- **WHEN** a fighter the bot was approaching steps behind a wall
- **THEN** the bot keeps going to the spot it last saw that fighter at, rather than stopping where it stood or tracking the fighter through the wall

#### Scenario: Giving up
- **WHEN** a bot reaches the remembered spot and nothing is in sight there
- **THEN** it waits a moment, then stops searching and goes back to standing and drifting its guard

#### Scenario: Patience runs out short of the spot
- **WHEN** the level stops a bot from ever reaching the spot it is searching for
- **THEN** the bot gives up after its patience is spent rather than pressing on indefinitely

#### Scenario: Sight cancels the search
- **WHEN** a bot part-way through a search sees a fighter
- **THEN** the search ends there and the bot reacts to the fighter it can see

#### Scenario: A searching bot does not swing
- **WHEN** a bot is walking to a remembered position
- **THEN** it does not attack, whatever it passes on the way

### Requirement: A bot gets over what it can and stops at what it cannot
A bot moving toward something SHALL read the level immediately in front of it and answer, each tick, whether it can keep walking, must leave the ground, or must stop.

- A **rise it can step onto** SHALL be walked onto, with no jump.
- A **wall it can clear** SHALL be jumped: a running jump when the bot is running and can afford one, a standing jump otherwise. Where the jump brings the bot to a ledge it can take hold of, it SHALL climb it as any fighter does.
- A **wall it cannot clear** SHALL stop the bot. It SHALL NOT stand pressed into the wall walking on the spot.
- A **gap it can cross** — one with ground on the far side within a jump — SHALL be jumped.
- A **gap it cannot cross** SHALL stop the bot at the near edge. A bot SHALL NOT walk off a ledge into a gap it has no way out of.

A bot SHALL jump only to get somewhere: it SHALL NOT jump while the way ahead is clear, and SHALL NOT jump at a wall or a gap it has already been found unable to clear.

None of this SHALL cost a bot anything a player would not pay. A running jump SHALL cost the bot its energy on the same terms, and a bot that cannot pay SHALL be treated as unable to make that jump rather than making it for free.

#### Scenario: A step is walked onto
- **WHEN** a bot walking toward a target meets a rise low enough to step onto
- **THEN** it walks up and keeps going, without jumping

#### Scenario: A wall is jumped and climbed
- **WHEN** a bot walking toward a target meets a wall within its jump
- **THEN** it jumps, catches the ledge and climbs onto it, and carries on toward the target

#### Scenario: An impassable wall stops it
- **WHEN** a bot meets a wall far taller than it can jump
- **THEN** it comes to a stop rather than walking into the wall indefinitely

#### Scenario: A crossable gap is jumped
- **WHEN** a bot moving toward a target meets a gap in the floor with ground on the far side within its jump
- **THEN** it jumps across and keeps going

#### Scenario: A bottomless gap is not walked into
- **WHEN** a bot moving toward a target meets a gap it cannot clear
- **THEN** it stops at the edge and does not walk off

#### Scenario: No jumping for its own sake
- **WHEN** a bot is walking toward a target across clear, flat ground
- **THEN** it does not jump

#### Scenario: A jump it cannot afford is a jump it cannot make
- **WHEN** a bot with too little energy for a running jump meets a gap only a running jump would cross
- **THEN** it stops at the edge instead, and may cross once it has the energy again

### Requirement: A bot follows a target off its own ledge
A bot whose target is on a different level of the ground SHALL steer toward the part of the level that would take it there — the near edge of the platform it must leave or reach — rather than walking into the wall or standing over the gap between them. A target below SHALL be dropped or jumped down to when the way down is safe by the rules above; a target above SHALL be climbed or jumped to when there is a way up within reach.

Where there is no way to the target that the bot can find by reading the ground in front of it, the bot SHALL come to a stop rather than press into the geometry, and SHALL react again as soon as the target is somewhere it can go.

#### Scenario: Following up
- **WHEN** a bot sees a target standing on a ledge above it, with a climbable edge in reach
- **THEN** it moves to that edge, jumps and climbs up, rather than walking on the spot beneath the target

#### Scenario: Following down
- **WHEN** a bot sees a target on ground below it, reachable by dropping down
- **THEN** it goes to the edge and drops down after it

#### Scenario: No route means standing still
- **WHEN** a bot sees a target it has no way to reach — across a gap too wide and with no way round it can read from where it stands
- **THEN** it stops rather than walking into the gap or into the wall, and moves again when the target comes somewhere reachable

### Requirement: Sight and movement decisions belong to the authority
Everything a bot sees and decides SHALL happen only where the match is actually simulated — the offline game and the dedicated server. A client predicting its own fighter SHALL NOT run any of it. What a client and its spectators see of a bot SHALL be exactly what they see of any other fighter: the movement, jumps and attacks it was told about, interpolated and drawn like anyone else's.

No new information about a bot's state of mind — what it can see, where it last saw someone, whether it is searching — SHALL be sent over the wire, and this change SHALL NOT alter the wire format or require clients and servers of the current protocol version to be rebuilt together.

#### Scenario: A client does not think for a bot
- **WHEN** a client is connected to a server with bots in the match
- **THEN** the client runs no sight or movement decisions for them; the bots move on the client purely from what the server sent

#### Scenario: Nothing new on the wire
- **WHEN** this change is in place
- **THEN** the packets exchanged between client and server are byte-for-byte the same as before for the same match
