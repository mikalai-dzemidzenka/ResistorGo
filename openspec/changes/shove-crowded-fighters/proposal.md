## Why

A fighter can end up standing with another on its back. With the reach
corrected (`symmetric-swing-reach`) that is a one-sided position: the fighter
behind can hit, the one in front cannot, and nothing in the game gets the
crowded fighter out of it except walking away or noticing in time to turn. A
shove answers it: the crowded fighter turns on whoever is on its back and throws
it off to arm's length, where the two face one another and either can swing.

`push.png` was drawn for this. A first version of the shove was built on the old
premise - two fighters standing in one another, neither able to reach - and was
withdrawn when the reach fix made that position impossible. This change brings
it back on the premise that is actually left.

## What Changes

- A new actor state, `StatePush`, appended to the state enum. A fighter that is
  standing still and has another overlapping it from behind, out of reach of its
  swing, turns to face that fighter and shoves it. It plays `push.png` over 7
  frames, cannot be steered, and lands on the frame both arms are out.
- Only a standing fighter shoves. A fighter running, attacking, blocking, in the
  air, flinching or guarding everything does not, so a chase - where the pursuer
  is always behind the pursued - never sets it off.
- The shove moves the crowder away from the shover, far enough to clear the
  overlap and near enough to stay inside the reach of a swing. It takes no
  structure, no energy, counts as no hit and does not interrupt what the crowder
  was doing.
- A hit landing on the shover during the shove interrupts it, as a hit
  interrupts anything a standing fighter does.
- The shove's displacement is the authority's alone, as a hit is.
- The reach test the bot uses is lifted out and shared with the shove, so what
  decides a shove and what decides a swing cannot drift apart.
- `push.png` is appended to the animation table and registered as a texture.

Depends on `symmetric-swing-reach`, which is committed. The shove fires where a
swing cannot reach, so it inherits that geometry exactly.

## Capabilities

### New Capabilities
- `shove`: what a standing fighter does about another on its back - when a shove
  fires, which way it goes, what it moves, what it costs, and who decides it.

### Modified Capabilities

<!-- none: no existing capability covers close-quarters positioning -->

## Impact

`actor.go` carries the state, its constants, the crowding test and the shove.
`bot.go`'s `inSwingRange` becomes a caller of the extracted reach test.
`resources.go` and the animation table gain `push.png`; both the state enum and
the table are append-only because their values travel in every snapshot, and the
decoder's bound already follows the enum's end.

Tests that stand two fighters near one another have to account for the shove:
`jump_test.go` spaces its fighters apart. Multiplayer: the server decides the
shove; the client is shown it by the snapshot.

A first implementation exists in the transcript of the session that built it,
on the old premise. It is the starting point, not the answer - the trigger, the
direction and the distance all change.
