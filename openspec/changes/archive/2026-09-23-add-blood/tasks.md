## 1. Droplets and stains

- [x] 1.1 Add `blood.go` with the constants, the droplet list, the stain ring
      and the blood's own `rand.Rand`; verify it builds
- [x] 1.2 Add `spray(at, dir, n)` and the droplet step (gravity, sub-steps of at
      most a pixel, landing against `LE.solids` as a stain stamped with a given
      tick, dropping droplets past their life or below the level); verify with
      tests on a floor and a wall that droplets land as stains on the floor,
      that some stain the wall when it is close behind, and that none remain in
      flight after `BloodLife`
- [x] 1.3 Cap both: verify that past `BloodDroplets` the oldest droplet goes,
      and past `BloodStains` the newest 4096 stains are kept in order
- [x] 1.4 Verify the blood's random source is its own: a fight with and without
      spraying draws the same numbers from the global source (done by
      parsing `blood.go` for any use of the global `rand`: since Go 1.24
      `rand.Seed` does nothing, so the global sequence cannot be replayed)

## 2. What spills blood

- [x] 2.1 Add `bloodWatch(w)`: per fighter, compare structure and state with
      what was last seen, spray by the rule in design.md away from the
      attacker, and remember the new values; verify `A flinch spills a little`
      and `More for a stun, most for a kill` with fighters driven through
      `takeHit`, and the direction by the side the droplets land on
- [x] 2.2 Verify `Blocked blows are clean`: a guarded block, block-all (its
      cost and a death from it), a clash and a shove spill nothing, and
      regeneration and a kill's reward spill nothing
- [x] 2.3 Verify `The match is unchanged`: the same scripted fight run with
      `bloodWatch` each frame and without gives identical fighters every tick

## 3. Time manipulation

- [x] 3.1 Draw only stains with `tick <= CL.Tick` while `TimeShowsHistory()`;
      clear droplets on entering the mode and skip spraying and flying while
      `matchHeld()`; verify with a test that restoring an earlier frame hides
      a later hit's stains and restoring the newest shows them again
- [x] 3.2 Delete stains newer than the frame on `rewindResumeHere` (seen by
      `updateBlood` on the first live frame after a recorded one, so no hook
      in `rewind.go`), and reseed the watch from the world on
      every restore and on leaving the mode; verify `Resuming keeps the
      frame's blood` and that a restore never reads as a hit
- [x] 3.3 Clear the blood in `startOffline`, `restartOffline` and `leaveMatch`;
      verify `A fresh level` (`clearBlood` is tested; the three callers need a
      window and are checked in play, 5.1)

## 4. Drawing and wiring

- [x] 4.1 Call `bloodWatch(CL)` and the droplet step once per rendered frame in
      `runLevel`, not on the dedicated server; draw stains after `LE.Draw()`
      and droplets after the fighters, both before `compositeLights()`, culled
      to the camera's view
- [x] 4.2 Run `go test ./...` and `go test -tags release ./...`; verify only the
      known flaky tests ever fail

## 6. Splashing onto the backdrop

- [x] 6.1 Keep an opacity mask for every texture in `loadTexBytes` (a bit a
      pixel, alpha over half), by resource name; verify with a test that
      decodes a small PNG with a transparent corner and checks the mask
- [x] 6.2 Give each droplet a reach when sprayed; once it has flown that long,
      stick it as a streak where the pixel under it is painted on a level
      tile, else fly on; verify with tests that a spray in front of a backdrop
      tile streaks the tile, that a spray over empty background leaves nothing
      in the air and stains the floor, and that a droplet over a tile's
      transparent pixel flies on
- [x] 6.3 Streaks: head plus a tail of 1-4 pixels back along the flight, drawn
      as a line darker toward the tail; verify with a test that a streak's tail
      points back along the droplet's velocity and is within the length, and
      that streaks count against the stain cap and follow time manipulation
      like stains

## 5. In the game

- [x] 5.1 Build with `build.ps1`; fight the bots in front of the backdrop
      wall: blood splashes onto the wall behind as streaks, flinches a little,
      stuns more, kills a burst; nothing hangs in the void; stains stay
- [x] 5.2 Fight in a dark corner and under a light: the blood is darker in the
      dark; stains are under the fighters and droplets over them
- [x] 5.3 Rewind across a kill and resume before it: the blood goes and comes
      back with the scrub, and stays gone after resuming
- [x] 5.4 Play a long fight past 4096 stains: the frame rate holds
- [x] 5.5 Play a networked match with a spectator: both see blood from each hit
- [x] 5.6 Tune the spray sizes, speeds and gravity by eye
