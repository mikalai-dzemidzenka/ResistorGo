## Context

`Cursor.Update(dt, energy)` is given the local fighter's energy as a fraction once per rendered frame by `runLevel`, and `DrawGUI` draws the crosses after the palette pass. A block's payback (`EnergyBlockRestore`, 15) is applied by `confirmBlock` in an authoritative world only; a client learns of it from a snapshot. Regeneration is predicted on both sides and adds at most `60 / EnergyFreeRegenTicks` points a second. A respawn creates a new `Actor`, so `w.Local` changes pointer; `TimeShowsHistory()` is true while a recorded frame is on screen. See proposal.md for why.

## Goals / Non-Goals

**Goals:**
- One detector for both flashes that works identically offline and online.
- Detection and the ring's shape over time as pure functions, tested without a window.

**Non-Goals:**
- A simulation event for the flash. It would fire on the server and in an offline world but never on a client, whose payback arrives in a snapshot.
- Flashing on other fighters or in the world; the flash belongs to the local crosshair.

## Decisions

**Watch the energy, not an event.** The cursor keeps the fraction it was given last frame. A rise larger than regeneration could explain in `dt` plus half a block's payback starts the small flash; reaching 1 from below starts the full one, and takes precedence. Half the payback, rather than all of it, still catches a block that is clipped by the cap — though that one fills the pool and shows the full flash anyway. Allowing for `dt` keeps a long frame's worth of refilling from reading as a block. Alternative: an `EvEnergyRestored` event bit beside `EvBlocked` — rejected, because on a client the payback never runs locally (the energy spec keeps it the server's alone), so the flash would only work offline.

**`runLevel` says whether to trust the change.** `Cur.Update(dt, energy, live bool)`, with `live` false when there is no local fighter, when `w.Local` is not the fighter it was last frame (kept in a package var beside `Cur`), or while `TimeShowsHistory()`. When not live, the cursor takes the new value as its baseline and starts nothing. This covers a respawn's jump to full and every jump a rewind makes, without the cursor knowing about actors or time.

**Two rings, one timer.** The cursor holds which flash is showing and its age; a new flash restarts it. `flashRing(kind, age)` returns the ring's inner and outer radius and colour, or nothing once the flash is over:

| | small (block) | full |
|---|---|---|
| lasts | 0.25 s | 0.5 s |
| radius | 10 → 26 px | 10 → 48 px |
| thickness | 2 px | 3 px |
| colour | white | bright green (180, 255, 190) |

The radius eases out (fast, then settling) and the alpha falls linearly to 0. The ring is centred on the attack cross as it moves. Ages advance by the frame's real `dt`, so a flash plays out through hit-stop and pauses like the crosshair itself does. Drawn with `rl.DrawRing` in window pixels, after the palette pass, so its colours are not quantized.

**Added, not modified.** The flash is a new requirement beside the crosshair ones; no existing requirement changes, so it cannot conflict with the in-flight energy changes.

## Risks / Trade-offs

- [A server correction that raises the local fighter's energy by more than half a payback flashes as if blocked] → The only upward corrections large enough are paybacks themselves; regeneration is predicted and replicated, so it does not drift that far.
- [At a very low frame rate a frame's refilling approaches the threshold] → The threshold grows with `dt`, so refilling alone never crosses it.
- [A block landing in the same frame as a respawn] → Not live that frame, so no flash; the fighter is new and full anyway.
