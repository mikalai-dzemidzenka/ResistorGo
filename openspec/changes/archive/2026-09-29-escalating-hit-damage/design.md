## Context

See proposal.md - Why. What shapes the approach is that the fighter already has a count of recent hits, and it is the wrong one.

`HitCount` is the progress toward the next stun. It is deliberately cleared every third hit — that is what "every third hit stuns" means — and it is replicated in exactly two bits, clamped on both sides of the wire to the largest value a stun leaves room for. A run that escalates has to survive the stun, so it cannot be that count, and it cannot fit in those bits.

The snapshot record is a fixed 44 bytes. Its flag word is a `uint16` carrying eight booleans in the low bits and the stun count in the two above them, leaving six bits unused. That is where a small count can go without the record growing — but any reinterpretation of those bits is a change an older build would read wrongly, and the handshake already refuses a client whose protocol version differs.

Structure is spent through one function that takes a cost and reports whether it emptied the pool. Nothing about that has to change; what changes is the number handed to it.

## Goals / Non-Goals

**Goals:**
- The escalation decided in one place, from one count, so the cost a hit will have is never computed two different ways.
- No growth in the snapshot: the run rides in bits the record already carries.
- The run ends everywhere the fiction says it ends — a block, recovery, death — and nowhere else.

**Non-Goals:**
- Changing when the third hit stuns, what a block restores, what block-all costs, or the size of the pool. The stun count and the run are separate counts that happen to move together for the first three hits.
- Escalating anything else: a block-all's cost and the structure a block restores are flat, and stay flat.
- Showing the run in the HUD. The bar already shows the consequence; a second row of pips for the cause is a separate question.

## Decisions

**The run is its own count, beside the stun count, not derived from it.** They diverge on the third hit and never line up again, so one cannot be recovered from the other. The alternative — reusing `HitCount` — costs nothing on the wire but caps escalation at three hits forever, because the stun that the count exists to trigger would also relieve the pressure. That is the opposite of what a run should do.

**The cost is computed from the run before the hit is counted, not after.** A hit's price is what the fighter has already taken, so the first hit of a run reads a run of zero and costs the base. Counting first and then pricing would make the opening hit of every run cost 1.5, which is a different game and a silent off-by-one that no single test would obviously catch.

**The count stops at the point the cost caps.** Once the run is long enough that the cost has reached its ceiling, a longer run means nothing, so the count stops climbing there. This keeps the value small enough for three bits and means the wire never has to carry a number whose only use is to be clamped — the clamp is the rule, not a transport limitation.

**The run rides in spare flag bits, and the protocol version moves.** Three bits above the stun count hold everything the run can be. Growing the record instead would cost every fighter in every snapshot for a value that fits in three bits, and snapshots are the one packet with a stated size budget. Reusing bits an older build reads as something else is exactly what the version is for: the handshake refuses the mismatch rather than letting two builds disagree about what a hit costs.

**Recovery ends the run, on the same trigger that already clears the stun count.** The requirement says pressure has to be sustained; the code already has one moment that means "this fighter has been left alone long enough" — the point where its structure starts refilling — and both counts clear there. Giving the run its own separate timer would be a second answer to a question already answered, and the two would drift.

**Both blocks end it, including block-all.** A block-all that stopped the attack is a stopped attack; the defender has already paid structure for the privilege of not having to read the height. Charging it the run as well would mean the panic button relieves nothing, which makes it not a button.

## Risks / Trade-offs

- **Fights get much shorter — six hits rather than twelve** → that is the intent, but it doubles the value of every opening and halves the room for error. The escalation step and the cap are each a single constant, so the curve can be retuned from one place after it has been played.
- **A client predicting its own fighter must agree with the server about the run, or it shows the wrong bar and corrects visibly** → the run is replicated for exactly this reason, and a scenario pins that a broken run reaches clients too, not only a growing one.
- **The version bump locks out older clients** → already true of every protocol change here, and the handshake refuses them with a reason rather than letting them misread a hit's cost. Worth saying out loud because this is the first change to reuse previously unused bits rather than add a field.
- **Two counts that look alike will be confused by a future reader** → they are named apart and commented at their declaration, and the one place that clears both says why it clears both.
