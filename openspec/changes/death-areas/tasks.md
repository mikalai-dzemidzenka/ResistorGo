## 1. Death areas in the level

- [ ] 1.1 Add `DeathArea{X, Y, Width, Height int32}` with the short JSON keys a
      lit area uses and a `rect()` method, in a new `death_areas.go`; verify
      with a test that one round-trips through JSON as `{"x":..,"y":..,"w":..,"h":..}`
- [ ] 1.2 Add the `deathAreas` key to `parseLevel` (one more return value) and
      to `marshalFile` behind `len > 0`, and load it into `le.deathAreas` in
      `LoadLevel`; verify with a test that a level with two death areas saves
      and loads them back unchanged, and that a level with none is saved
      with no `deathAreas` key
- [ ] 1.3 Change `marshal` to a `map[string]any` with the same layer keys plus
      `deathAreas` when there are some; verify the `A level without them keeps
      its shape` scenario (loading and saving `level_1.json` without an edit
      gives the same bytes and the same `Hash()` as before the change) and
      `Client and server disagree on a death area` (adding one changes `Hash()`)

## 2. Touching one kills

- [ ] 2.1 At the end of `Actor.Update`, on the authority only and for a fighter
      not already in `StateDeath`, call `enterDeathState(nil)` when
      `collisionRect()` overlaps any of `LE.deathAreas`; verify `Falling into
      a pit` and `One pixel is enough` with a test that drops a fighter into
      an area and checks the tick it dies on, and that a one-pixel overlap kills
- [ ] 2.2 Verify `Nobody gets the kill`: a fighter last hit by another dies in
      an area, the other's energy and structure are unchanged, and the script
      death event carries killer id 0
- [ ] 2.3 Verify `A body is left alone` and `Nothing to stand on`: a dying
      fighter passing through an area is not killed again, and a fighter
      walking beside an area it doesn't touch moves exactly as it does with
      no area
- [ ] 2.4 Verify `The client waits to be told`: following
      `net_collision_rects_test.go`, a networked test where the client's own
      fighter walks into a death area. The predicted world does not kill it,
      the server does, and the client shows the death from the snapshot

## 3. The death tool

- [ ] 3.1 Add `toolDeath` after `toolArea`, and `deathAreas`, `deathSel` and
      `deathDrag` to `LevelEditor` (`deathSel` starts at -1); verify that Q
      cycles place → rect → light → area → death → place, and that the toolbar
      shows five toggles
- [ ] 3.2 Add the death tool's glue in `death_areas.go`, copying the area
      tool's: `deathAt`, `placeDeath` (taking the selected one's size, or
      `defaultArea`'s), `beginDeathDrag`, `dragDeath`, `setDeath`,
      `deselectDeath`, `deleteDeath` and `updateDeathTool`, plus the
      `toolDeath` case in the editor's update and Delete handling; verify by
      copying the area tool's editor tests to cover placing, moving and
      resizing, never inside out, removing, letting go, and `Tools keep to
      their own`
- [ ] 3.3 Add the death tool's panel: position and size fields for the
      selected area, or the count and how to place one; verify `Typing a death
      area exactly`, `Nothing selected`, and that digits typed into a field
      neither retool the editor nor delete anything
- [ ] 3.4 Add death areas to the undo snapshot in `level_undo.go`, cloned and
      restored like `areas`, with `deathSel` clamped; verify `Undone like any
      edit`, and that selecting and deselecting leave the level saved

## 4. Seeing them

- [ ] 4.1 Draw every death area in a palette red outline over `markerBacking`,
      with the selected one filled faintly, while the death tool is out; draw
      the same outline in the debug overlay beside the collision boxes, and
      nothing with debug mode off. Verify by eye at the closest and farthest
      zoom, with the death tool out, with another tool out in debug play, and
      with debug mode off
- [ ] 4.2 Verify death areas never reach `computeLights`: the level is lit the
      same with and without one, by eye and by checking that
      `computeLights`'s callers only pass `le.areas`

## 5. Verify in play

- [ ] 5.1 Run `go test ./...` and `go test -tags release ./...` and confirm
      both pass
- [ ] 5.2 Build with `./build.ps1` and play `level_1.json`: put a death area in
      the pit under the drop-through plate, drop into it, and confirm the
      fighter dies, flickers and respawns; save, restart and confirm the area
      is still there and still kills
- [ ] 5.3 Play a networked match against the dedicated server with the same
      level: walk into the death area under `F5`-`F8` latency simulation, and
      confirm the fighter dies once, on both screens, with no correction spike
      and no death shown on the client before the server's
