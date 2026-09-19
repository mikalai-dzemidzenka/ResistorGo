## Why

Two fighters can end up standing in one another - a bot's lunges walk it into
whoever it is fighting, a player can walk into a bot, a chase closes the gap -
and nothing in the game parts them. Overlapping fighters read as a glitch, and
the one behind can swing at the other's back while the other cannot swing back.
A shove parts them: whichever is on its feet and free turns on the other and
throws it off to arm's length, stunned but unhurt, the two facing one another.

`push.png` was drawn for this. A first version of the shove was built for a
narrower position - two fighters in one another, neither able to reach - and
withdrawn when the reach fix made that position impossible. It came back first
for a fighter on another's back only; in play, overlaps in front were left
standing, so the trigger is now any overlap.

## What Changes

- A new actor state, `StatePush`, appended to the state enum. When two fighters
  on the ground overlap, in front, behind or on the same spot, the first one on
  its feet and free - standing, walking or running - turns to face the other and
  shoves it. It plays `push.png` over 7 frames, cannot be steered, and lands on
  the frame both arms are out.
- A fighter swinging, blocking, stunned, flinching from a hit, in the air or
  guarding everything does not shove; it finishes what it is doing. A fighter in
  the air is not shoved until it lands.
- The shove moves the crowder away from the shover, far enough to clear the
  overlap and near enough to stay inside the reach of a swing, and stuns it from
  the moment the shove begins - which also keeps it from shoving back. It takes
  no structure, no energy and counts as no hit. Without the stun the crowder -
  usually mid-swing - cut nearly every shove off in play.
- A hit landing on the shover during the shove interrupts it, as a hit
  interrupts anything a standing fighter does.
- The shove's displacement and stun are the authority's alone, as a hit is.
- The reach test the bot uses is lifted out into one function, so the shove's
  distance bounds and the bot's stopping distance measure the same reach.
- `push.png` is appended to the animation table and registered as a texture.

Depends on `symmetric-swing-reach`, which is committed: the shove's distance is
bounded by that reach.

## Capabilities

### New Capabilities
- `shove`: what two overlapping fighters do about it - who shoves, which way it
  goes, how far, what it costs, and who decides it.

### Modified Capabilities

<!-- none: no existing capability covers close-quarters positioning -->

## Impact

`actor.go` carries the state, its constants, the crowding test and the shove.
`bot.go`'s `inSwingRange` becomes a caller of the extracted reach test.
`resources.go` and the animation table gain `push.png`; both the state enum and
the table are append-only because their values travel in every snapshot, and the
decoder's bound already follows the enum's end.

Tests that stand two fighters near one another have to account for the shove:
`jump_test.go` spaces its runners apart, and the bot attack test updates its
target so the stun a bot's own shove leaves it in can wear off. Multiplayer: the
server decides the shove; the client is shown it by the snapshot.

The first implementation is kept beside this change in `recovered.md`, as the
starting point it was.
