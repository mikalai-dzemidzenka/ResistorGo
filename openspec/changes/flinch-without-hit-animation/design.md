## Context

See proposal.md - Why. What shapes the approach is where the flash was decided from.

The white flash was never a property of the fighter alone. The renderer drew a fighter through the flash shader when `AnimWhite` was set *and* the current animation frame was below 2 — so the flash was really "the opening of the hit animation", and it worked only because every hit set that animation. Take the animation away from the flinch and the second half of that test starts reading a frame number of some unrelated animation: a fighter hit while walking would flash or not depending on where its walk cycle happened to be.

The two ends of the wire also have to agree. A fighter's flash is decided on the machine drawing it, from whatever that machine holds: the server's own actor, the client's predicted local fighter, or a remote fighter rebuilt from snapshots. Only replicated fields are available in the last case, and `FlinchTicks` is not one of them. The snapshot record is a fixed 44 bytes with two spare flag bits; adding the hold to it is possible but is a protocol version bump for something the receiver can already work out.

## Goals / Non-Goals

**Goals:**
- One predicate, asked by every draw path, that decides the flash from state a client is guaranteed to hold.
- The flash cannot outlive what raised it, on any path out of a flinch.

**Non-Goals:**
- Changing what a hit costs, how far it throws, or when it stuns — see the `structure` spec.
- Growing the snapshot. No new replicated field, no protocol version bump.
- A flinch animation of any kind, now or later behind a flag.

## Decisions

**The flash is asked of the actor, not derived at the draw site.** The renderer asks one predicate and draws through the shader if it says so. Alternative: leave the condition inline and add the flinch case to it — rejected because the condition is now two rules (a stun flashes over the opening of its animation; a flinch flashes for its hold), and the spectator, replay and any future draw path would each have to carry the same pair.

**The stun case is told apart by the fighter's state, not by its hold.** A stun is exactly the fighters in the hit-stun state, which is replicated; everything else flashing is a flinch. So the predicate reads: not flashing unless the flash flag is set; in the hit-stun state, flash over the opening frames of the animation; otherwise flash. Alternative: replicate `FlinchTicks` in the two spare flag bits (it fits in four) and test it directly — rejected as a protocol change buying a distinction the state already makes.

**The hold stays where it is and keeps its job.** The flinch hold already existed and already cleared the flash when it ran out; it loses only its second job of suppressing animation changes while the hit animation played, which is what the idle branch that froze a flinching fighter's animation was for. That branch goes, which is also what makes "SHALL NOT interrupt an action the defender had already started" true rather than nearly true.

**Clearing the flash belongs with clearing the hold.** The old code zeroed the hold in several places without touching the flash, and got away with it because the frame test would go false on its own once the new animation played past frame 2. With the frame test gone, a missed clear leaves a fighter white for the rest of the match. So the two are cleared together by one call, used everywhere the hold was being zeroed — starting a swing, starting a block — with the stun and death paths setting the flash explicitly afterwards as they already did.

## Risks / Trade-offs

- **A future draw path tests the flash flag directly and gets the flinch case wrong** → the flag alone is no longer the answer; the predicate is. Nothing else reads the flag for display, and a test covers a remote-shaped fighter.
- **A new way out of a flinch forgets to clear the flash** → the clear is one call rather than an assignment to remember, and a test asserts a fighter that swings out of a flinch is not left flashing.
- **The flinch reads as weaker than it did** — it is a flash and a shove where it used to be a whole animation. That is the point of the change, but it does mean the hit-stop freeze is now the main thing selling the impact of the first two hits; if they end up reading as too light, the answer is the freeze or the knockback, not giving the animation back.
- **The stun's flash is shorter than its animation** → deliberate, so the flash marks impact and the animation marks the lockout; it does mean the two hit reactions differ in more than one way at once, which is what makes them readable.
