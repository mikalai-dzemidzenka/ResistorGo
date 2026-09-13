## Why

A flinch and a stun looked the same. Every landed hit played `hit-top.png`, so the two hits that only flinch wore the same 0.3-second reaction as the third one that actually takes the fighter's control away — and the flinch animation talked over whatever the fighter was doing, while the rule that a flinch "SHALL NOT interrupt an action the defender had already started" says it should not. A player could not read from the reaction whether they had been stunned or were still free to answer, which is the one thing the third hit is supposed to announce.

Telling them apart costs nothing: the hit flash is already there, and the stun already owns an animation whose end is what releases the lockout. Giving the animation to the stun alone makes the reaction mean something, and leaves a flinch as what it is — a flash, a shove and a dent in the structure bar.

## What Changes

- A flinch no longer plays the hit animation. The fighter keeps the animation it was already playing — walking, running, swinging, falling — and is shown to have been hit by the white flash alone, held for about 0.2 seconds.
- The stunning third hit still plays the hit animation, and is now the only hit that does. Its flash covers the opening of that animation rather than the whole of it, so the flash still reads as the moment of impact and the lockout that follows is the animation.
- Everything else a hit does is unchanged: 1 structure, the shared hit-stop freeze, the knockback (small for a flinch, larger for a stun), the stun count and the bar.
- The flash ends with whatever raised it. A flinch's flash stops when its hold runs out; a fighter that swings, blocks, is stunned or dies before then takes the flash with it, so nothing can leave a fighter lit white.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `structure`: *A landed hit costs structure and flinches* — the flinch no longer includes the hit animation, and the flash is given a duration of its own. *Every third landed hit stuns* — the stun is named as the only hit that plays the hit animation.

## Impact

- `actor.go` — `hitFeedback` no longer sets an animation; `enterAttackedState` plays `hit-top.png`; `enterFlinchState` sets only the flicker hold; `stopFlicker` ends a flinch that something else interrupts; `Flickering` decides the white pass for the renderer. The `StateIdle` branch that froze a flinching fighter's animation is gone.
- `game.go` — the draw loop asks `Flickering()` instead of testing `AnimWhite` and the animation frame itself.
- Nothing is added to the wire: the flash rides on the already-replicated `AnimWhite`, state and animation fields, so a client and the server decide it the same way.
- No change to the resources: `hit-top.png` is still loaded and still used by the stun.

This behavior is already in the tree — it shipped in commit `9f84d10`, ahead of this change being written, which is why its tasks are weighted toward verifying what is there rather than building it.
