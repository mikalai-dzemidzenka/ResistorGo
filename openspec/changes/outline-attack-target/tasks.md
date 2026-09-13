## 1. One attack box, asked for by both sides

- [x] 1.1 Add a method on `Actor` that returns the attack box for a given guard height, built from a position the caller supplies: the three existing offsets (`AttackBoxLow`, `AttackBoxMid`, `AttackBoxHigh`), the same facing term, and no rectangle at all for guard height 3. Verify `go build ./...` succeeds.
- [x] 1.2 Have `updateAttack` use it in place of its inline `switch`, passing the simulated position. Verify the combat tests still pass unchanged — `TestOneLandingPerSwing`, `TestEveryThirdHitStuns` and the block tests in `structure_test.go` — so the geometry is identical to what it replaced.
- [x] 1.3 Add a test that the method returns the three heights at the offsets the fighter carries, mirrored when it faces left, and nothing for block-all. Verify it fails if any offset is changed.

## 2. Which fighter is marked

- [x] 2.1 Add a predicate in a new `attack_outline.go` taking the aiming fighter and a candidate, answering whether that candidate is outlined: a local fighter exists and is not the candidate, neither is in the Death state, the aimer's guard height is an attack height, time is not showing history, and the aimer's attack box overlaps the candidate's collision box. Verify by reading that it touches no field a client does not hold.
- [x] 2.2 Add tests for the range rule: a candidate just inside the box is marked and one just outside is not, a candidate reachable at one height only is marked at that height only, and a candidate behind the aimer is not marked — the `Walking into range`, `Height changes what is in range` and `Facing away` scenarios. Verify each fails if the overlap test is dropped.
- [x] 2.3 Add tests for the exclusions: block-all marks nothing, a dead candidate is not marked, a recorded frame marks nothing, the aimer never marks itself, and a nil local fighter marks nothing — the block-all, `A dying target`, `Scrubbing time`, `Never itself` and `No fighter to aim with` scenarios. Verify they pass.
- [x] 2.4 Add a test that two fighters both inside the box at the current height are both marked, the `Two in range` scenario. Verify it passes.

## 3. The band

- [x] 3.1 Add a layout function returning the outlined band in whole world pixels: the attack box's vertical extent at the aimer's height, across the marked fighter's sprite frame, positioned from the marked fighter's `drawPos()` so it tracks the interpolated sprite rather than the simulated one. Verify every edge is a whole number.
- [x] 3.2 Add a test that the band's height follows the aimer's guard height — low below mid below high — and that it covers the marked fighter's body, the `The band sits at the attack's height` scenario. Verify it passes.
- [x] 3.3 Add a test that the band is laid out from the interpolated position: with a marked fighter's simulated and interpolated positions apart, the band follows the interpolated one, so the outline cannot lag the sprite. Verify it fails if the layout reads `Pos` instead.
- [x] 3.4 Add a test that the band is identical whatever the marked fighter's guard height, structure, energy and hit count are — the requirement that the outline says nothing about defence. Verify it passes.

## 4. Drawing it

- [x] 4.1 Add an outline shader that traces a sprite frame's silhouette: discard the fighter's own pixels, keep a transparent pixel only where the body touches it, discard everything outside the band, and clamp neighbour sampling to the frame's own columns so a neighbouring pose cannot bleed in. Verify `go build ./...` succeeds.
- [x] 4.2 Add one place that reports the texture, source rectangle and position the fighter's sprite is drawn with, and draw the outline from it so the two passes cannot drift. Verify by reading that it matches what `DrawAnimOnce` and `DrawAnim` use, including the frame and the facing.
- [x] 4.3 Add the draw function: convert the band into the sprite's own coordinates, set the shader's uniforms, and draw the marked fighter's sprite a second time through it, in a red taken from the palette so the palette pass leaves it unchanged at full strength. Verify it draws nothing when the band misses the sprite entirely.
- [x] 4.4 Fade the outline across the band in the shader: full strength along the middle row and nothing at the top and bottom edges, from the band's unclipped range so the fade stays centred on the row the swing lands on when the band runs past the sprite. Verify `go build ./...` succeeds.
- [x] 4.5 Call it from the overlay loop in `runLevel`, before the structure bars and after the lighting, over the same sorted draw list. Verify `go build ./...` succeeds and that the lighting cannot tint it.
- [ ] 4.6 Verify the debug overlay is untouched: its three red attack boxes still draw in debug builds, and `go build -tags release ./...` still succeeds with the outline present.

## 5. Whole-change verification

- [ ] 5.1 Run `go test ./...` and `go test -tags release ./...` and confirm both pass, the combat and replication suites included.
- [ ] 5.2 Play a local match (`go run ./cmd`): confirm the band slides up and down the bot as the aim moves, appears as the bot comes into reach and goes as it leaves, and that block-all clears it.
- [ ] 5.3 Confirm in play that the outline tracks a moving target without lag at several camera zooms, and that its edges stay on the pixel grid.
- [ ] 5.4 Play a networked match against the dedicated server (`go run ./cmd -dedicated` plus a client): confirm each player sees the other outlined and never itself, that it holds under `F5`–`F8` latency simulation, and that a spectator sees no outlines at all.
- [ ] 5.5 Confirm in play that the outline reveals nothing about the defence: watch a target change guard height under the band and confirm the band does not react.
