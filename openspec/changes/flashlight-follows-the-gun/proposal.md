## Why

The pistol's flashlight shines from every living fighter holding the pistol,
but the pistol is only drawn while the fighter stands or walks on its pistol
legs. Climbing a ledge, jumping, rolling or taking a hit, an armed fighter
shows no gun while the beam goes on shining from where the muzzle would be.

## What Changes

- The flashlight shines only while the fighter's pistol is drawn. It goes out
  on the climb, the climb-down, a jump or fall, a duck or roll and a hit, and
  comes back when the fighter is on its pistol legs again.
- The setting (options screen, F3) and the beam itself are unchanged.

## Capabilities

### New Capabilities

### Modified Capabilities
- `firearms`: the flashlight follows the drawn pistol.

## Impact

- `flashlight.go`: `flashlightBeams` skips a fighter whose pistol is not
  drawn, by the same test `topAnim` uses to draw it.
- `gun_test.go`: `TestFlashlightBeams` arms its fighter the way play does, and
  gains a climbing case.
- Presentation only: nothing is replicated, so no protocol or prediction change.
