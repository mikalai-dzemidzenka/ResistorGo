## 1. Replicate the hit count

- [x] 1.1 In `net_protocol.go`, pack `HitCount` into bits 8–9 of `ActorRec.Flags` in `recOf` and read it back in `apply`, clamped to `StunEveryHits-1`; raise `ProtocolVersion` to 4 and leave `actorRecSize` at 42. Verify with a unit test that hit counts 0, 1 and 2 survive `recOf`/`apply`, that `TestActorRecSize` still passes unchanged, and that the flag bits already in use are unaffected.
- [x] 1.2 Verify the "Stun progress is replicated" scenario with a unit test: a fighter's record after two `takeHit` calls carries 2, and after the third (the stun) carries 0.

## 2. Bar logic, without a window

- [x] 2.1 Add `showStructureBar(a *Actor) bool`: structure below `MaxStructure`, not `StateDeath`, and `TimeShowsHistory()` false. Verify with unit tests covering full structure (hidden), one hit (shown), the Death state (hidden), time manipulation showing a recorded frame (hidden), and regeneration back to full (hidden again).
- [x] 2.2 Add a geometry helper that returns the bar's track, the two fill halves and the pip squares in whole world pixels, from `drawPos()`, the collision box, `Structure` and `HitCount`, per design.md (16 × 2 bar, 3 px above the collision box; each fill half `round(8 × lost/12)`; up to two 2 × 2 pips, 2 px apart, in a row centred 1 px above the bar). Verify with unit tests: at 9 structure the fill is 4 px and symmetric about the bar's centre, at 1 structure it is nearly full, a block from 8 to 8.5 shrinks it, every rectangle has integer coordinates even when the fighter's interpolated position is fractional, the pip count is clamped to two, and one or two pips sit above the bar, within its width, centred on it.
- [x] 2.3 Define the track, fill and pip colours as exact `Palette` entries. Verify with a unit test that each of the three is present in `Palette`, so the palette pass cannot alter it.

## 3. Draw it

- [x] 3.1 Add the drawing function and the extra pass in `runLevel`: after the additive light pass and before `EndTextureMode`, reopen `BeginMode2D(Cam.View())`, walk the draw list, and draw each fighter's bar where `showStructureBar` allows — independent of `Debug`. Verify the game builds, that a bar appears over a fighter on its first hit and follows it while it runs, and that it disappears once regeneration refills the pool.
- [x] 3.2 Verify in game that the pips count up with the first and second hits and clear on the stun, while the bar stays.

## 4. Verify against the spec

- [x] 4.1 Run `go test ./` and confirm the whole suite passes, including the new protocol and bar tests.
- [x] 4.2 Play through the remaining scenarios in `specs/structure-bar/spec.md` that need a running game: the bar zooms with the camera and stays on the pixel grid; two fighters at the same structure, one under a lamp and one in shadow, have bars of identical colour (compare the pixels in a screenshot); no bar while a fighter is dying; no bars while time manipulation shows a recorded frame, and they return when it is left; bars still shown with the debug overlay off.
- [x] 4.3 Verify the remote-fighter scenario with an integration test in the style of `TestNetDeathAndRespawn`: after the server lands two hits on a fighter, the client's copy of that fighter carries the server's structure and a hit count of two.
