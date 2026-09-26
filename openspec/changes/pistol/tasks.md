## 1. Sheets and gun math

- [x] 1.1 Append `resources/threefour/gun-bot-idle.png` and `resources/threefour/pistol-top.png` to the end of `animTable` in `resources.go`, and verify `go test ./...` still passes (the animation-id tests) and both textures load when the game starts
- [x] 1.2 Create `gun.go` with the gun constants (`WeaponPistol`, `ShotRange` 240, `ShotCooldownTicks` 15, `FlashSec`, `TracerSec`), the per-frame angle and muzzle table from design.md, `gunAim(from, cursor, facing)` (byte from cursor), `aimDeg(aim)`, `aimFrame(aim)` (nearest frame) and `muzzle(a)` (mirrored when facing left); verify with `gun_test.go` asserts that level, highest and lowest cursors give frames 2, 0 and 4, that out-of-range cursors clamp, and that the muzzle mirrors to `47 - x`
- [x] 1.3 Add `rayRect(origin, dir, r)` (slab test, entry distance) and `shotEnd(a, w)` (nearest of non-drop-through solids and living non-ally fighters other than the shooter, else full range); verify with tests: a wall in front of an enemy stops the shot at the wall, the nearer of two enemies is hit, an ally and a dead fighter are passed, a drop-through platform is passed, and an empty line ends at 240 px

## 2. Holding the pistol

- [x] 2.1 Add `Weapon`, `Aim`, `Shots`, `ShotCooldown` (simulation) and `SeenShots` (presentation only) to `Actor`, and `Gun`, `Aim` to `Input`; latch Q in `latchInput` into `Input.Gun`, clear it in `clearLatchedInput`, and fill `Input.Aim` in `sampleInput` from the shoulder to the mouse; verify with `go build ./...`
- [x] 2.2 Handle `Input.Gun` in the Idle and Run states on the ground: toggle `Weapon`, drop Run to Idle when drawing, and move `Updown` 3 to 1 when drawing; verify with tests that Q while standing draws, Q again holsters, Q while jumping does nothing (and nothing on landing), and Q while running leaves the fighter in Idle holding the pistol
- [x] 2.3 Add the Idle gun branch: copy `Input.Aim` into `a.Aim`, set the body to `gun-bot-idle.png` frame 0, skip updown, walk and block animations and swings, and keep jump, roll and climb down; clamp speed to `RunReach` in `updateHorSpeed` while armed; verify with tests that an armed fighter holding a move key for 2 s never enters `StateRun` and that jumping then landing returns it to the gun legs with `Weapon` still set
- [x] 2.4 Add `&& actor.Weapon == 0` to the melee `guarding` check in `updateAttack`; verify with a test that a high swing at an armed fighter whose input is high lands as a hit, not a block
- [ ] 2.5 Extend `topAnim`/`topDraw` so an armed fighter on `gun-bot-idle.png` gets `pistol-top.png` at frame `aimFrame(a.Aim)`; verify in play, looking right and left, that the torso sits on the legs, follows the mouse through all five directions, and the rim light traces it

## 3. Firing

- [x] 3.1 Add `fire()`: require `ShotCooldown == 0`, set the cooldown, increment `Shots`, and on the authority resolve the `shotEnd` target (stunned or clashing: nothing; blocking: `exitBlockState` then `takeHit`; else `takeHit`); count `ShotCooldown` down each live tick; call it from the Idle gun branch on `Input.Attack`; verify with tests that a shot costs the target a first hit's structure and flinches it, that block-all does not stop it, that it kills and credits the shooter at empty structure, that a stunned target is unchanged, that a second click within 15 ticks does not fire (then or later), and that the shooter's energy is unchanged
- [x] 3.2 Verify with a test that a non-authority world firing a shot increments `Shots` but changes no target's structure

## 4. Tracer and flash

- [x] 4.1 In `gun.go`, keep presentation lists of tracers and flashes aged by `dt`; in `presentEffects` spawn one of each when `a.Shots != a.SeenShots` (end from `shotEnd`) and then sync `SeenShots`; set `SeenShots = Shots` where the client first creates a fighter from a snapshot; verify with a test that stepping `presentEffects` after one shot spawns exactly one tracer, that a second call spawns none, and that both expire after their lifetimes
- [ ] 4.2 Pass the flashes as extra `LevelLight`s to `computeLights` in `game.go` (without mutating `LE.lights`), and draw tracers after `compositeLights()` beside the rim pass; verify in play that a shot in a dark part of the level briefly lights it and casts shadows, and that the tracer ends at a wall or at the fighter hit

## 5. Multiplayer

- [x] 5.1 Protocol: `Gun` in command button bit 7 and an `Aim` byte after `RollAim`; `Weapon`, `Aim`, `Shots`, `ShotCooldown` in `ActorRec`, with `recOf`, `apply` (unknown `Weapon` → 0) and the encoder and decoder, `actorRecSize` 51; `ProtocolVersion` 13; verify with round-trip tests for both and the existing 16-fighter size test staying at or below 1200 bytes
- [x] 5.2 In `Server.runCmd`, also rewind for a command that fires (armed and `in.Attack`); verify with a lag-comp test that a shot fired at a target in the aim at the view tick hits after the target has moved out on the server, and that every target's position is restored afterwards
- [x] 5.3 Verify with a client prediction test that a reconcile replaying a shot tick leaves `Shots` at the predicted value and spawns no second tracer, and with a server and client integration test that a client's Q then LMB draws and fires on the server

## 6. Smoke and dust

- [x] 6.1 In `gun.go`, add puffs (smoke at the muzzle for every shown shot; dust at the end point when the shot stopped at a solid) with drag, lift, growth and a lifetime, aged by `ageShots`; verify with a test that a shot at a wall spawns smoke and dust, a shot at a fighter spawns smoke only, puffs rise over their life, and all are gone after it
- [ ] 6.2 Draw puffs as dithered pixel discs inside the lit scene, beside `drawDroplets` in `game.go`; verify in play that smoke drifts up from the muzzle and thins out, that dust comes off a wall, and that the flash lights the smoke

## 7. Gun feel

- [x] 7.1 Stop a landed shot from freezing the shooter; verify with a test that after a shot lands the target is frozen and the shooter is not
- [ ] 7.2 Draw the recoil kick from `ShotCooldown` in `topDraw` (one direction up, or a pixel back at the top) and add a local camera kick; verify with a test that the drawn torso frame is one above the aim for the first ticks after a shot and back at the aim after, and by eye that the view jolts on your own shot
- [ ] 7.3 Keep the shot direction on the flash and draw the muzzle fire in `drawTracers`; verify by eye
- [x] 7.4 Add casings and sparks as bouncing, falling particles and bullet holes capped at 64, all cleared by `clearShots` wherever `clearBlood` runs; verify with tests that a wall shot spawns sparks, a hole and a casing, a fighter shot spawns a casing only, a casing dropped over the floor comes to rest on it, holes are capped, and everything is gone after its life
- [ ] 7.5 Verify in play that casings bounce and rest, sparks fly off walls, holes stay and fade out, and the muzzle fire reads at the barrel

## 8. Sound

- [ ] 8.1 Load `pistol1.mp3` and `pistol2.mp3` as a voiced sound set beside the punches and play one for every shot shown; verify with a test that a shown shot is heard exactly once (through an `onShot` hook), and by ear in play

## 9. Wrap-up

- [ ] 9.1 Run `go test ./...` and verify it passes, then build with `build.ps1` and check in play (offline, and host plus join) that drawing, aiming, walking, firing, the tracer, the flash, kill credit, the smoke, the wall dust, the gunshot sound and respawning unarmed all behave as specified
