## Context

Two things in the code already look like a death area:

- **Lit areas** (`LitArea`, `level_lights.go`) are freestanding rectangles of the level with their own editor tool (`toolArea`), their own key in the level file (`litAreas`), and undo. Their editing is thin glue around shared rectangle helpers: `rectHandleAt`, `handleAnchor`, `dragRect` and `rectDrag`.
- **The Death state** is entered through `enterDeathState(attacker)` (`actor.go`), which already treats a nil attacker as "no killer": the killer's reward is skipped, and `scriptEvent` sends the scripts killer id 0. There is no kill feed to update.

The level fingerprint (`LevelEditor.Hash`) covers the tiles only (`marshal`): lights, lit areas and the ambient are left out because they don't change the simulation. A death area does change it, so it has to be in the fingerprint.

## Goals / Non-Goals

**Goals:**
- A death area kills through the existing Death state, with no new state, animation or message.
- The editor's death tool works exactly as the area tool does, so there is nothing new to learn.
- Levels without death areas are untouched: same file bytes, same fingerprint.

**Non-Goals:**
- Making the area editing generic over both kinds of area. Only the thin glue is duplicated (see Decisions).
- Bot awareness. `bot.go`'s gap and edge checks don't see death areas.
- Changing `parseLevel`'s return shape beyond one more value.

## Decisions

**A separate type and tool, not a kill flag on `LitArea`.** A kill flag would reuse every line of the area tool, but it would put a gameplay rectangle in a lighting list, with a feather and a darkness it would have to ignore, and the lighting pass would have to filter it out. The user chose the separate type. `DeathArea{X, Y, Width, Height int32}` uses the same short JSON keys a lit area does (`x`, `y`, `w`, `h`), with a `rect()` method.

**Duplicate the glue; don't generalise the area tool.** The code that's hard to get right is already shared: hit-testing handles, the grab offset, dragging without turning inside out. What's specific to lit areas is about ten lines per function: `areaAt`, `placeArea`, `beginAreaDrag`, `dragArea`, `setArea`, `deselectArea`, `deleteArea`, `updateAreaTool` and the drawing. The death tool gets its own copies (`deathAt`, `placeDeath` and so on) in a new `death_areas.go`, beside `DeathArea` and the kill check. The alternative, a generic editor over `[]T` with `rect()`/`withRect()` methods, would mean rewriting a tool that works and is covered by tests to save about 60 lines.

**The fifth tool goes after the area tool.** `toolDeath` joins the enum before `toolCount`. The toolbar sizes itself from `toolCount` and Q cycles with `nextTool`, so that is all the wiring the tool switch needs. The panel, the Delete key and the update switch each get a `toolDeath` case, following `toolArea`.

**Level file: key `deathAreas`, written only when there are some.** The key is added to the recognised keys in `parseLevel` and written in `marshalFile` behind `len > 0`, exactly as `litAreas` is. `parseLevel` returns one more value. That's six, which is clumsy but a local cost, and replacing it with a struct is a refactor this change doesn't need.

**Fingerprint: `marshal` covers death areas when there are some.** `marshal` builds `map[int]any` from the layers. It becomes `map[string]any` with the same layer keys, plus `"deathAreas"` when the level has any. `encoding/json` writes map keys sorted as strings in both cases, so a level without death areas produces the same bytes and the same hash. A test pins this against `level_1.json`.

**Where the kill happens: at the end of the fighter's tick, on the authority only.** In `Actor.Update`, after movement and collision have moved the fighter for the tick, a check runs when `a.W != nil && a.W.Authority && a.State != StateDeath`. If `collisionRect()` overlaps any of the level's death areas, it calls `enterDeathState(nil)`. Checking after movement means a fighter falling fast still dies on the first tick its box touches the area. The death areas are read from `LE`, as `dropThrough` reads `LE.drop`, which is how every piece of level geometry already reaches the simulation on both the client and the server (`LE.LoadLevel` in `game.go` and `net_server.go`).

The alternative was a world-level pass after all actors have updated. That would work too, but it would need its own loop and ordering relative to `reap` and `updateRespawns`, and the per-actor check sits where the fighter's position becomes final.

**Drawing.** While the death tool is out, the death areas are drawn in a red outline, with the selected one filled faintly, using `markerBacking` under the outline so they stay legible over light and dark, as `drawAreas` does. The debug overlay draws the same outline next to the collision boxes. Nothing is drawn with debug mode off, and death areas are never passed to `computeLights`.

**Undo.** The undo snapshot in `level_undo.go` gets a `death []DeathArea` field, cloned and restored like `areas`, and restoring clamps `deathSel` like `areaSel`.

## Risks / Trade-offs

- [A bot walks into a death area it can't see] → accepted for now and named in the proposal. Placing death areas where the course bots run is up to the level designer until pathing learns about them.
- [The duplicated glue drifts from the area tool's behaviour] → the specs mirror the area tool's scenarios one for one, and the editor tests for the death tool are written by copying the area tool's.
- [The fingerprint change breaks matchmaking between builds] → only for levels that have death areas. `level_1.json` keeps its hash until someone places one, and a test pins that.
- [Rewind across a death-area death] → death is an ordinary state in the rewind ring, so rewinding past it restores the living fighter like any other death. If the rewound fighter is still overlapping the area, it dies again on the next tick, which is correct.
- [A fighter respawns inside a death area] → `spawnPos` knows nothing about death areas. The level designer keeps spawn points clear of them. If one is placed badly, the fighter dies on spawn and respawns again, which is visible and not a crash.

## Open Questions

- The default size of a newly placed death area when none is selected. `defaultArea` is 64x64 and will do; tune it in the editor.
- The exact red for the markers, chosen from the palette so it survives the palette pass.
