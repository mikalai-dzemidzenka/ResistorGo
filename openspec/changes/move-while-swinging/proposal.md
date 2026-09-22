## Why

A fighter that starts a swing is planted: movement input is ignored for the
whole attack, and the only ground it covers is the 4 px dash each blow opens
with. So once two fighters are trading, where they stand is decided by who
swung first, not by footwork - a fighter cannot ease out of an incoming swing's
reach while its own is still playing, or ease in to make its own land. Letting
a swinging fighter move, very slowly, makes distance something both fighters
work at through an exchange.

"Very slowly" is not possible today. Every tick the horizontal position is cut
to a whole pixel, so any speed under 1 px a tick - 60 px/s - moves nothing at
all. The slow creep this needs is under that floor, and so is the flinch's
knockback, which is why a flinch moves nobody (the open
`flinch-knockback-is-lost-to-rounding` change). Both have the same fix.

## What Changes

- Horizontal position keeps its fraction from tick to tick instead of being
  cut to a whole pixel; it is rounded only where it is drawn, as it already
  is. Vertical position is unchanged. Every horizontal speed now moves the
  fighter by the distance it describes.
- A fighter in an attack may move left or right at up to 20 px/s - about 6 px
  over a low swing - while holding its facing: it can back away from, or
  press toward, the fighter it faces without turning. Bots move the same way
  as players when their input asks them to.
- Walking becomes the same everywhere. Cutting the position toward zero was
  never neutral: moving away from x = 0 lost the pixel's fraction every tick
  and moving toward it gained one, so a one-second walk today covers 78 px to
  136 px depending on the direction and which side of the origin the fighter
  is on, and the slide after letting go 26 px to 60 px. Without the cut, the
  same walk covers the same distance anywhere in the level. This is a fix, and
  it changes how far walks go: some faster than today, some slower.
- Letting go of the keys stops a walking fighter once it drops under 60 px/s
  with no key held, trimming about 7 px of the slide the old floor hid and
  landing near today's average. Knockbacks are not stopped this way.
- A flinch's knockback now moves the fighter, a few pixels, clearly less than
  a stun's throw. This settles `flinch-knockback-is-lost-to-rounding` in
  favour of its second option; that change is superseded.

## Capabilities

### New Capabilities
- `movement`: how a fighter's horizontal position advances - sub-pixel
  position, moving while swinging, and stopping when the keys are let go.

### Modified Capabilities
- `structure`: a flinch's knockback must actually move the fighter, and stay
  shorter than a stun's throw.

## Impact

`actor.go`: the truncation after `Pos.X += Speed.X * dt` goes; `updateHorSpeed`
admits movement during `StateAttack` at a capped speed without turning, and
gains the stop for a released walk; the attack case's clamp of `Speed.X` to
`AttackDash` (4 px/s, a speed that could never move anyone) becomes the new
cap. Collision snapping, the attack dash, the shove's pixel glide and the
vertical position are untouched.

Everything that moves under 60 px/s starts moving, so the ends of the
knockbacks lengthen: the stun's measured 20 px throw and the shove's distance
need re-measuring. Tests that assert exact whole-pixel positions after
movement may need their expectations updated. Networking and time manipulation
already carry positions as floats and copy whole fighters, and need no change.
The spec's convention that positions are integers each physics step is
changed for the horizontal axis.
