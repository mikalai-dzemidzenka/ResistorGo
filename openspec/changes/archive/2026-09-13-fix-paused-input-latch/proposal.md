## Why

Clicking while time manipulation holds the match makes the fighter attack the moment the match resumes. The attack button is latched for the next fixed step (`pressedAttack`), but while the mode is on no fixed step runs, so the latch survives until the mode is left and is then consumed by the first step back. The `time-manipulation` spec forbids exactly this, in three separate places: fighter input "SHALL NOT be applied ... until the mode is left", "no attack is queued or resolved", and "no input is buffered for when the mode is left". The code does not match the spec it already has.

This is the second time the same bug has appeared. Edit mode had it too, and was fixed there by hand with its own gate and its own clearing line. The two pause paths each maintain their own copy of "is the match held?", spelled differently:

```
latchInput   is gated on   !spectating && !Editing
stepOffline  is gated on   TimeMode == TimeOff && !Editing
```

Neither list mentions the other's condition, which is how time manipulation came to be missing from the first one.

## What Changes

- **One rule for "the match is held".** A single predicate answers it for every pause path — edit mode and time manipulation today, anything added later — and both the input latch and the fixed-step loop are gated on it. The two hand-maintained condition lists go away.
- **One place that drops stale input.** Leaving either paused mode clears the latched attack and jump, through a single shared helper rather than a line copied into each mode.
- **Clicking while time is paused no longer attacks on resume.** That is the user-visible fix; edit mode keeps the behaviour it already has.

### Assumptions

- Movement and stance are read live from the keyboard and mouse inside the fixed step, so they are already suppressed by the step not running. Only the latched attack and jump need clearing.
- Spectators keep their existing exclusion from the input latch; the shared predicate is about the match being held, which is a different question from having no fighter.
- No timing or gameplay value changes. A click made before a pause begins is still dropped when the pause ends, exactly as edit mode already does.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `time-manipulation`: *Gameplay input is suppressed while the mode is on* gains an explicit scenario for the resume case, mirroring the one `dev-tools` already has for edit mode. The requirement's wording already forbids the behaviour; the new scenario names it directly so it is covered by a test rather than implied.

## Impact

- **Code**
  - `actor.go`: the shared `clearLatchedInput` helper beside `latchInput`.
  - `game.go`: the shared "match is held" predicate, and `runLevel` gating both the latch and `stepOffline` on it.
  - `rewind.go`: leaving time manipulation clears the latch.
  - `level_editor.go`: edit mode's hand-written clearing line becomes a call to the shared helper.
- **Tests**: a regression test that a click latched while time manipulation is on does not survive the mode being left, mirroring the existing edit-mode test.
- **Not affected**: the scrubbing rules, the recording, edit mode's behaviour, gameplay values, networking and the menu.

## Non-goals

- Changing what time manipulation or edit mode do beyond this input rule.
- Reworking how input is latched or sampled.
- The attack zoom punch never returning from maximum zoom, which is a separate bug in the same area.
