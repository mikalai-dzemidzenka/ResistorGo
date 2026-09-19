## Why

A fighter's swing reaches 38px in front of it facing right and 32px facing left.
The attack box is mirrored about the centre of the 48px sprite frame (24), but
every reach test compares it against the collision box, which is centred at 21.
The 3px offset flips sense when a fighter turns around, so the same swing covers
a different distance depending only on which way the fighter happens to be
looking.

This has been true since the attack box was written and passed unnoticed while
reach only ever decided whether a hit landed. The shove (`StatePush`) made it
load-bearing: the shove fires when boxes overlap and the swing cannot reach, and
at dx = 0 a fighter facing right cannot reach while one facing left can. A wall
clamp puts two fighters at exactly dx = 0, which produces a permanent soft-lock
against right-hand walls - both fighters shove forever, and neither can move,
jump or attack again for the rest of the match.

## What Changes

- **BREAKING (gameplay)**: the swing's reach becomes the same in both facings.
  The attack box is mirrored about the collision box's centre instead of the
  sprite frame's centre. Maximum reach becomes 35px both ways, in place of the
  current 38 forward / 32 backward split - 3px shorter facing right, 3px longer
  facing left.
- A fighter standing at exactly dx = 0 is reachable in both facings, where today
  only a left-facing fighter can reach one.
- The crowding window that triggers the shove becomes symmetric: `(-14, -3]`
  facing right and `[3, 14)` facing left, in place of `(-14, 0]` and `[6, 14)`.
- Bot approach distances (`inSwingRange`, `tooClose`) are re-checked against the
  new reach, since they were tuned while the two facings disagreed.

## Capabilities

### New Capabilities
- `swing-reach`: how far a swing reaches, measured the same way in both facings,
  and the rule that a fighter standing on another can be hit rather than only
  shoved.

### Modified Capabilities

<!-- none: no existing capability under openspec/specs/ covers attack reach -->

## Impact

`attackBoxFacing` in `actor.go` is the single chokepoint, and it is load-bearing
in five places:

- `attackBoxAt` -> `updateAttack` - where hits land
- `attackBoxAt` -> `attack_outline.go` - the drawn target outline moves with it
- `attackBoxAt` -> `net_lagcomp.go` - server-side rewound hit tests
- `swingReachFacing` -> `inSwingRange` (`bot.go`) - where a bot stops walking
- `swingReachFacing` -> `crowding` (`actor.go`) - what triggers the shove

Tests that pin reach or bot approach distance will move: `bot_course_test.go`,
`push_test.go`. Multiplayer is affected on both sides - clients predict hits with
the same box the server confirms them with, so client and server must ship the
same constant.

Not in scope: the sprite art is drawn 1px left of its frame centre, so a
fighter's drawn body shifts 2px sideways on turnaround while its collision box
does not. That is an art-registration issue, tracked separately.
