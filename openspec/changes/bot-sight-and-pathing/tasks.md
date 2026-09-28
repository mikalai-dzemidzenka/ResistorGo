## 1. Move the bot out of actor.go

- [x] 1.1 Create `bot.go` and move `botInput` and `Actor.nearest` into it unchanged; verify `go build ./...` and `go test ./...` still pass and `bot_guard_test.go` is unaffected
- [x] 1.2 Create `bot_test.go` with a helper that builds an authoritative world on a hand-made `tiles(...)` layout and returns a bot and a target at given positions; verify it compiles and one smoke test using it passes

## 2. Sight

- [x] 2.1 Add the sight tunables — `SightRange`, `SightSlope` (tangent of the half-angle), `SightSearchTicks`, `SightLingerTicks` — beside the other bot tunables, each with a one-line comment saying what it means in play; verify `go vet ./...` is clean
- [x] 2.2 Implement the segment-versus-`LE.solids` slab sweep (`lineBlocked`), guarding the zero-component divisions; verify with tests covering a clear line, a line through one solid, a line that passes just above a solid, a vertical line and a horizontal line
- [x] 2.3 Implement `canSee(a, b)` — facing sign, squared range, slope cone, then `lineBlocked` — measured between collision-box centres; verify with tests for in front and clear, behind, out of range, outside the cone above, and blocked by a wall (spec: *A bot acts only on fighters it can see*)
- [x] 2.4 Implement `seen(a)` returning the nearest visible fighter, or nil; verify a test with two visible fighters picks the nearer and a test with one visible and one closer-but-hidden picks the visible one
- [x] 2.5 Rewire `botInput` to use `seen` in place of `nearest`, and to call `face()` only toward a fighter it can see; verify a test that a bot with nobody in sight produces an `Input` with no movement, no attack and no change of `Orientation`, and that its guard still re-rolls over many ticks (spec: *A bot that sees no one does not fight*)
- [x] 2.6 Verify committed actions survive losing sight: tests that a swing started on the tick before sight is lost still resolves, and that a bot hit by an unseen attacker still takes the hit, loses structure and does not turn (spec: same requirement)

## 3. Terrain probes

- [x] 3.1 Derive the jump-reach constants (`JumpRise`, `JumpRun`, `RunJumpRise`, `RunJumpRun`, plus the safety margin) from `Gravity`, `MaxJumpSpeed`, `MaxRunJumpSpeed`, `MaxSpeed` and `FixedTimeStep`; verify with a test asserting they land within a few pixels of a simulated jump arc, so retuning the physics cannot silently break the bot's belief about it
- [x] 3.2 Implement `wallAhead(a)` — the highest wall top within the lookahead window in the facing direction, using the same wall-versus-step rule as `Actor.isWall`; verify with tests over `tiles(...)` layouts: nothing ahead, a step-height rise, a one-tile wall, a wall taller than any jump
- [x] 3.3 Implement `groundAhead(a, x)` — the highest solid top under a probe column within the drop window, and "no ground" for a gap; verify with tests: flat floor, the tile after a missing one, and the column past the end of the floor
- [x] 3.4 Implement `gapAcross(a)` — given a gap ahead, the near edge and the first ground on the far side, and whether it is inside the standing or the run-jump reach; verify with tests for a one-tile gap, a gap only a run-jump crosses, and a gap wider than any jump

## 4. The step decision

- [x] 4.1 Define the `step` intent (`dir`, `jump`, `stop`) and implement `botStep(a, goal)` following the ladder in specs/bots/spec.md — step, clearable wall, unclearable wall, crossable gap, uncrossable gap — with `canRunJump()` consulted before choosing any jump that needs the run-jump's reach; verify with direct unit tests per branch on hand-built layouts (spec: *A bot gets over what it can and stops at what it cannot*)
- [x] 4.2 Handle a goal on another level of the ground in `botStep`: steer toward the near edge of the platform that leads up or down, and answer `stop` when neither is readable from where the bot stands; verify with tests for a target on a ledge above with a climbable edge in reach, a target on ground below reachable by dropping, and a target across an uncrossable gap (spec: *A bot follows a target off its own ledge*)
- [x] 4.3 Wire `botStep` into `botInput`: its `dir` becomes `Left`/`Right` and the `face()` call, its `jump` becomes `Input.Jump`, its `stop` produces no movement; verify a test that a bot walking at a one-tile wall ends up on top of it within a bounded number of ticks, and one that a bot at an uncrossable gap never leaves the ground (spec: same two requirements)
- [x] 4.4 Verify the bot does not jump for no reason: a test that a bot crossing clear flat ground toward a visible target sets `Input.Jump` on no tick

## 5. Search

- [x] 5.1 Add the non-replicated bot memory to `Actor` as one small grouped field (last seen position, patience counter), with a comment saying it never reaches the wire; verify `go test ./...` passes and no snapshot test changes
- [x] 5.2 Implement the search in `botInput`: losing sight records the position and starts the patience countdown, the goal handed to `botStep` becomes the remembered position, arriving switches to the linger, either counter reaching zero clears the memory, and a searching bot never attacks; verify with tests for walking to the last known spot, giving up after arriving, giving up when the spot is unreachable, and no attack while searching (spec: *A bot searches where it last saw someone*)
- [x] 5.3 Verify sight cancels a search: a test that a bot part-way to a remembered spot switches to the fighter it now sees and that the memory is cleared
- [x] 5.4 Verify a respawned bot starts with no memory: a test that a bot killed mid-search comes back neither searching nor holding the old position

## 6. Authority and the wire

- [x] 6.1 Verify nothing new is replicated: a test that serialising an actor mid-search produces the same bytes as the same actor with no memory, and confirm by inspection that `writeActor`/`readActor` and `ProtocolVersion` are untouched (spec: *Sight and movement decisions belong to the authority*)
- [x] 6.2 Verify a client thinks for nobody: a test over the existing net integration harness that a predicted client world runs no bot decisions, and that bots on a client move only from snapshots

## 6b. Alerted by an attack

- [x] 6b.1 Add `Actor.alerted(attacker)` — marks the attacker's position as the spot last seen, bots only — and call it from `takeHit` and `confirmBlock`; verify tests that a blow from behind turns the bot and that it then sees and fights its attacker (spec: *An attack tells a bot where its attacker is*)
- [x] 6b.2 Make a searching bot turn toward the remembered spot even when there is nowhere to step, so a blow landed from point-blank behind still spins it round; verify by the turn test above
- [x] 6b.3 Verify the limits: a blocked attack alerts the same, a hit does not make the attacker visible (it searches where the blow came from and gives up), a fighter already in view outranks a blow from behind, and a human is never steered

## 6c. Gaps are crossed at a run

- [x] 6c.1 Found in play: a bot that reached a crossable gap slowly stopped a stride short and never moved again — from a standstill it cannot predict enough speed at the edge, and a standing jump cannot cross a gap at all, because its 42-tick wind-up walks the fighter off the edge before take-off
- [x] 6c.2 Judge a gap by the running jump alone: it must be affordable, and the speed predicted at the edge must be a run. Where the far side is within a run-up jump but the room is not there, back away from the edge by `botRunUp` (derived from `RunReach`, `DefaultAcc` and the run-jump wind-up) and come at it again; verify `TestBotBacksOffToTakeARunAtAGap`
- [x] 6c.3 Never back off into a second hole; verify `TestBotDoesNotBackOffIntoAnotherGap`

## 6d. Losing the player at a wall

- [x] 6d.1 Draw the sight cone, the line to what a bot sees and the spot it is searching, in debug mode; after `compositeLights` so the light map cannot dim the wash, and only where the sight is real (the authority, never a client's predicted world)
- [x] 6d.2 Found in play with the overlay on: following a fighter that hops a ledge, the bot lost sight at the wall's foot — the spot last seen is then the ground it already stands on, so the search arrived instantly, lingered and gave up, leaving it blind against the wall for good. A stalled search now takes the top of a climbable wall in front of it as the next place to look; verify `TestABotLooksOverTheWallItLostSomeoneBehind`
- [x] 6d.3 Nothing within a body's width counts as behind, so two fighters on the same spot are not blind to each other; verify `TestFightersOnTopOfEachOtherAreNotBlind`
- [x] 6d.4 Read the ground ahead while airborne too, and hard-brake by turning when there is not even room to coast; verify `TestBotLandingOnAShortStripDoesNotRunOffIt` and `TestBotHardBrakesWhenCoastingWouldNotStopIt`

## 6e. The strip, and what widening it found

- [x] 6e.1 Widen the test course's strip below the high ledge from one block to four (304px): a bot coming off the ledge at a run needs about 205px to pull up, and with 76px it could only refuse the drop altogether
- [x] 6e.2 Measure holes from the column the fighter is actually held up by — the middle of its collision box, which is the one `checkCollision` looks under — not the leading face half a body ahead of it. Probing from the face, a bot lost the hole the moment the face crossed it and walked the last pixels unbraked, ending with its centre one pixel past the edge, stationary and falling
- [x] 6e.3 A hole is an empty column, not a high one: `edgeAhead` now asks whether there is anything at all to stand on, because reading a step up as a hole made a bot brake at the foot of a staircase and, in the air with no wall check ahead of it, brake backwards off the ledge it had just left
- [x] 6e.4 Look down off a ledge as well as up over a wall when a search runs out of ground; verify the end-to-end follow test
- [x] 6e.5 Add `TestABotFollowsAFighterTheLengthOfTheCourse` — a fighter walks the course and a bot follows it through every obstacle. Driving a bot at a fixed goal never reproduces one that loses sight of something moving, which was the whole failure; run it 25 times over for flakiness, the odds being unseeded

## 6f. A body is not a target

Found after the fact: `seen()` filtered on the sight cone alone, so a fighter
that had gone down was still the nearest thing a bot could see. The bot walked
to the body and swung at it until the world got round to removing it, and the
living fighter standing past it was never chosen.

- [x] 6f.1 Skip the dead in `seen()`, so a body is neither approached nor swung
      at; verify the `A body is not a target` and `The living are found past the
      dead` scenarios
- [x] 6f.2 Verify the course test still measures pathing rather than a kill: a
      bot that catches the leader and fights it long enough kills it, and with
      the dead no longer seen there is nothing left to follow and nothing the
      rest of the course can say. Hold the bot's swing back in
      `TestABotFollowsAFighterTheLengthOfTheCourse`, leaving everything that gets
      it along the course - the jumps included - as it is
- [x] 6f.3 Confirm nothing else treats a body as a live fighter: check the
      attack-target outline, the bot's guard adaptation and the spectator's
      follow target against a dead actor

## 7. Verify in play

- [x] 7.1 Run the full suite with `go test ./...` and confirm every existing test — combat, energy, structure, net, spectator, menu — still passes
- [x] 7.2 Give the game somewhere to show this off. `level_1.json` turned out to hold four collision rectangles and no floor at all — fighters spawn and fall for ever — so rather than author a level inside the one being built, this is a separate `level_bot_test.json` reached with a new `-level` flag: a course of a step, a climbable wall, a crossable gap, stairs to a high ledge, a drop and a gap no jump crosses. Verified by tests that load the file off disk and walk a bot along every obstacle
- [ ] 7.3 Play the offline game (`go run ./cmd -level level_bot_test.json`) and confirm by eye: the bot ignores you until you step into its cone, follows you to where it last saw you and gives up, jumps the wall and the gap, and stops at the edge of one it cannot cross
- [ ] 7.4 Tune `SightRange`, `SightSlope`, `SightSearchTicks` and `SightLingerTicks` by eye in play until the bot reads as alert rather than blind or omniscient
- [x] 7.5 Rebuild the release binary with `./build.ps1` (never a plain `go build`) and confirm the packed `resistor.exe` is replaced and runs
