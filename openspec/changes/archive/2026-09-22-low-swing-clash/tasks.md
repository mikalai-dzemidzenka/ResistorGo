## 1. The state and its assets

- [x] 1.1 Append `StateClash` to the `ActorState` enum before `stateCount`;
      append `resources/threefour/battle-low-1-left.png` and
      `battle-low-1-right.png` to `animTable` and register both in
      `LoadTextures`; verify `TestAnimTableIsAppendOnly` (pin the previous last
      entries, hold the two sheets to the end) and
      `TestEveryStateSurvivesTheWire` pass
- [x] 1.2 Add `ClashWindowTicks`, `ClashStructure`, `ClashSeconds` and
      `ClashGap` with the starting values from design.md, and the `SwingTick`
      field set from `a.W.Tick` in `enterAttackState`; verify with a test that
      an opening swing stamps the world's tick and a chained blow leaves it
- [x] 1.3 Add the `StateClash` case to `Actor.Update` (speed zeroed, input
      ignored, idle when the animation finishes) and make `face` refuse a turn
      in it; verify with a test that a clashing fighter given movement,
      attack and jump input stays put, does not turn, and is idle with no
      attack queued when the animation ends

## 2. Starting a clash

- [x] 2.1 Add `startClash(a, b)`: pay (swing energy for a fighter coming from
      a block, `ClashStructure` for both, death with the other as killer on a
      break and the survivor to idle), place by position at `ClashGap` through
      `blockSides`, face each other, clear queued and buffered input, set the
      two animations on the same tick; verify the `Both pay`, `Not a hit`, `A
      clash can kill` and `Sides follow position` scenarios, and that a clash
      against a wall leaves neither fighter inside it
- [x] 2.2 Add `tryClash(a, b)` with every condition from design.md; verify
      with table tests for each condition that fails it (allies, back to
      back, out of reach, mid or high or mixed heights, chained blow, outside
      the window, already clashing, stunned, dead, in the air, cannot pay from
      a block) and the ones that pass
- [x] 2.3 Call `tryClash` from the top of `updateAttack`'s authority branch,
      before the hit loop; verify the `Same tick` and `Inside the window`
      scenarios by driving two fighters through `Update`, in both update
      orders
- [x] 2.4 Call `tryClash(a.Attacker, a)` from the top of `updateBlock` when
      the blocker presses attack; verify the `Caught from the block` scenario,
      and that a press after the window leaves the block to play out
- [x] 2.5 Verify the `Outside the window`, `Other heights do not clash`, `A
      chained blow does not clash` and `Back to back does not clash`
      scenarios resolve exactly as before this change (the same hit or block)

## 3. Out of everyone else's reach

- [x] 3.1 Skip `StateClash` fighters in `updateAttack`'s target loop and in
      `crowding`; verify the `A third fighter swings at a clash` scenario
      (no hit, no block, no provisional block, structure unchanged) and that a
      clashing fighter overlapping a third is neither shoved nor shoves
- [x] 3.2 Verify `Protection ends with the clash`: a third fighter's swing
      landing just after the clash ends hits as usual
- [x] 3.3 Verify `Ending together`: both clashing fighters leave the state on
      the same tick

## 4. Authority and the wire

- [x] 4.1 Verify with a client-world test that a non-authority world never
      starts a clash from the same inputs
- [x] 4.2 Verify the `Online clash` scenario with the existing server/client
      test harness: a server clash reaches the client's own fighter through
      reconcile, with the server's structure and energy, and the replayed
      commands do not swing out of it
- [x] 4.3 Run `go test ./...` and `go vet`; verify nothing else changed
      behaviour

## 5. In the game

- [x] 5.1 Build with `build.ps1`; in single player, stand at arm's length from
      a bot and swing low as it swings low; tune `ClashGap` until the punch
      and the knee meet, and `ClashSeconds` until the exchange reads at speed
- [x] 5.2 Try the window by hand: a clash when pressing just after a bot's low
      swing starts, none when noticeably late; tune `ClashWindowTicks` if it
      feels off
- [x] 5.3 With two bots on the player, clash with one while the other swings;
      verify its swing passes through and neither fighter flashes
