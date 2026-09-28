## Context

See proposal.md for the reasons behind this change. Most of the data model already exists. `LevelEditor.objects` is `map[int]LevelLayer`. `LevelObject.Layer` is saved in the block table. `place` and `erase` already work on `le.layer`, and `selectAt` already picks the topmost layer in a cell. `buildIndex` sorts `objectsIndexed` by layer, then by Y descending, then by X descending, and builds `solids` from every layer. What's missing: nothing writes `le.layer`, and `game.go` calls `LE.Draw()` once, before everything else, with a `TODO sort objects with actors`.

## Goals / Non-Goals

**Goals:**
- Draw tiles in front of fighters.
- Choose the layer in the editor.
- Load old levels without editing them by hand.

**Non-Goals:**
- Choosing a layer per texture (a default layer stored for each palette entry).
- Hiding the other layers while editing. (They are faded while the place tool is out, so the current layer stands out; see the editor tasks.)
- Depth-sorting tiles against fighters within one layer.
- Parallax.
- Collision that depends on layer.

## Decisions

**Layer 0 holds fighters only.** The layer stepper skips 0, and the draw pass treats any layer `<= 0` as background. Tiles only reach layer 0 through code that builds `LE.objects[0]` directly, which is what tests do. The alternative was to sort tiles and fighters against each other on a shared layer 0. That ordering means nothing in a side view, since fighters are sorted by X, so the TODO is removed rather than implemented.

**Split one draw pass into two instead of adding a draw list.** `objectsIndexed` is already sorted by layer, so `Draw` takes a predicate or range: a pass for `Layer <= 0` and a pass for `Layer > 0`, walking the same slice. The debug and editor overlays (rectangle outlines, grid, cursor, unbuilt and missing markers, selection) move to the foreground pass, so they stay on top of foreground tiles. A merged tile-and-actor draw list was rejected because it would only matter if tiles and fighters shared a layer.

**Order in `runLevel`:** background tiles, then stains and holes, then fighters, then droplets, casings and puffs, then foreground tiles, then `compositeLights`. Foreground tiles go before `compositeLights`, so lights and ambient light shade them.

**Migrate on load, in `parseLevel`.** Before placements are decoded into layers, the `"0"` key's placements go to `-1` and each tile's `Layer` becomes -1. Blocks carry the layer, so it has to be rewritten on the decoded objects and not only on the map key. Client, server and editor all load through `parseLevel`, so their fingerprints stay equal. The alternative was a one-time manual rewrite of `level_1.json`. It would fix only that one file, and the "is this file migrated" question would still be open for every other level.

**PageUp/PageDown step the layer.** Neither key is bound in the editor. Stepping has no upper or lower limit, because a layer is just an `int`. The panel shows `layer: N` next to the tool status.

## Risks / Trade-offs

- [Rim light shows through foreground tiles, because it is added after `compositeLights`.] → Accepted for now; see the proposal. The fix is to mask the rim with foreground coverage, and it's worth doing once there's real foreground art.
- [Fingerprint changes on the first save after migration.] → Client and server builds already ship together. Both migrate the same way on load, so even an unsaved old file matches.
- [A background tile placed on -1 in a cell that holds another -1 tile replaces it, as today.] → Unchanged behaviour. Stacking needs a different layer.
- [Blood stains are drawn under foreground tiles.] → This is correct: the tile is in front.

## Migration Plan

Load and save `level_1.json` once in the editor, then commit the rewritten file with the rebuilt exe. Rolling back means reverting both.
