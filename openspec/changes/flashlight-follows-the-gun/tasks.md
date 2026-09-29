## 1. Beam follows the drawn pistol

- [x] 1.1 In `gun_test.go`, make `TestFlashlightBeams` arm its fighter with
      `toggleGun` (so it is on its pistol legs, as in play) and add a case: the
      same armed fighter in the climb (`enterClimbState`) casts no beam. Verify
      the climb case fails today
- [x] 1.2 In `flashlightBeams`, skip a fighter whose pistol is not drawn - the
      `a.Anim == gunLegs` test `topAnim` draws the pistol by; verify
      `TestFlashlightBeams` passes and `go test ./...` passes

## 2. Verify in the real game

- [ ] 2.1 Build with `.\build.ps1` and play with the flashlight on (F3): draw
      the pistol, then climb a ledge, jump, roll and take a hit. No beam shows
      without the gun, and it comes back each time the fighter stands armed
