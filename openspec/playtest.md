# Playtest backlog

Every playtest and tuning task left open when these changes were archived on
2026-09-29. Tags are `change task`; the change now lives under
`openspec/changes/archive/2026-09-29-<change>/`, and "record in design.md"
means that change's design.md there. Tick here, not in the archived tasks.md.

## Code debt

- [x] `horizontal-collision 3.1` Following `net_collision_rects_test.go`, add a
      networked test: a floor with a wall, the client's own fighter held running
      into it. It stops at the same X on client and server, and `c.Corrections`
      is zero (the *Online against a wall* scenario).
- [ ] `bot-sight-and-pathing 6f.2` Hold the bot's swing back in
      `TestABotFollowsAFighterTheLengthOfTheCourse` so it measures pathing, not a
      kill; leave the jumps as they are.
- [ ] `bot-sight-and-pathing 6f.3` Check the attack-target outline, the bot's
      guard adaptation and the spectator's follow target against a dead actor.

## Builds

- [ ] `horizontal-collision 4.2` `./build.ps1 -Out <scratch>` and
      `./build.ps1 -Linux` both pack.
- [ ] `main-menu-is-a-terminal 6.5` `.\build.ps1`, run it, and the menu matches
      the development build.
- [ ] `roll-breaks-the-surround 3.2` Run the full suite and the course-following
      bot test several times; nothing regresses.

## Session 1: offline (`go run ./cmd`)

### Menu
- [ ] `main-menu-is-a-terminal 6.1` Keyboard only on the main screen: W/S and the
      arrows wrap at both ends, Enter and Space take the row, the highlight
      follows the mouse the moment it moves and the two are never both lit.
- [ ] `main-menu-is-a-terminal 6.2` Escape on Host Game with an error shown, on
      Join Game mid-edit, on Options after moving the volume, and on Connecting:
      each leaves as its Back row does, and no field comes back still editing.
- [ ] `neo-minsk-main-menu 4.4` Tune the cycle speed and ramp entries by eye;
      record changed values in design.md.

### Movement
- [ ] `add-run-jump-and-fall 6.2` A standing jump still plays the wind-up, leaves
      the ground when it finishes, and reaches the run-jump's height.
- [ ] `add-run-jump-and-fall 6.4` Run and jump: the energy bar drops by half, a
      third jump on an empty pool is refused with the bar flashing, take-off is
      immediate, the running-jump animation plays, the fighter steers in the
      air, the fall animation takes over at the apex, and landing at speed
      continues the run.
- [ ] `add-run-jump-and-fall 6.5` Walking off a ledge falls, a ledge caught on the
      way down still climbs, and a fighter knocked off an edge falls without
      disturbing the hit reaction.
- [ ] `roll-breaks-the-surround 3.2` Held from both sides, the roll escapes.

### Taking hits
- [ ] `escalating-hit-damage 5.1` A run of hits drops the bar further each time,
      six in a row ends it, and one block puts the next hit back to a small step.
- [ ] `escalating-hit-damage 5.2` Block-all breaks a run as well as a matching
      guard, and so does walking away long enough to start recovering.
- [ ] `flinch-without-hit-animation 4.1` A hit while running: the run continues
      and the fighter flashes. The third hit plays the hit animation and takes
      control away.
- [ ] `flinch-without-hit-animation 4.2` The flash never sticks: swing, block and
      jump out of a flinch, and die during one; the fighter draws normally after.
- [ ] `stun-throws-back 4.1` A third hit visibly throws the fighter, further than
      the two flinches, and reads as the exchange resetting.
- [ ] `stun-throws-back 4.2` Judge the 20 px throw: stun or nudge? If wrong,
      change `StunKnockback` and record value and reason in design.md.
- [ ] `stun-throws-back 4.3` Judge the overrun (still creeping ~0.06 s after
      control returns): off balance, or a lost moment of recovery? Record the
      verdict in design.md; stopping it is a follow-up change.
- [ ] `stun-throws-back 4.4` Stun a fighter with its back to a wall and one at a
      ledge: it stops or falls as a flinched fighter does.

### Hit-stop input
- [ ] `fix-hitstop-input-loss 5.1` Take a hit and click during the freeze: the
      counter comes out, and clicking a moment later behaves the same.
- [ ] `fix-hitstop-input-loss 5.2` Land a swing and mash through the hit-stop:
      the chain follows as reliably as on a whiff.
- [ ] `fix-hitstop-input-loss 5.3` Take a hit and press W during the freeze: the
      fighter leaves the ground as it ends.
- [ ] `fix-hitstop-input-loss 5.4` The third hit of a run, click during the
      freeze: nothing comes out during or at the end of the stun.
- [ ] `fix-hitstop-input-loss 5.5` Chains with no hit: the click still registers
      from anywhere in the swing.

### Attacking
- [ ] `symmetric-swing-reach 4.1` The forward swing still feels right 3 px
      shorter, and the backward swing no longer feels short.
- [ ] `symmetric-swing-reach 4.2` Debug mode: the attack-box outline is mirror-
      symmetric about the collision box when a fighter turns.
- [ ] `outline-attack-target 4.6` Debug mode still draws the three red attack
      boxes.
- [ ] `outline-attack-target 5.2` The band slides up and down the bot with the
      aim, appears as it comes into reach and goes as it leaves; block-all
      clears it.
- [ ] `outline-attack-target 5.3` The outline tracks a moving target without lag
      at several zooms, edges on the pixel grid.
- [ ] `outline-attack-target 5.5` The band does not react when the target
      changes guard height.

### Bots
- [ ] `bot-adapts-guard-after-hit 4.1` Hitting the bot at one height makes it
      guard there some of the time, and the attack stops being reliable.
- [ ] `bot-adapts-guard-after-hit 4.2` Not every hit: the guard sometimes stays.
- [ ] `bot-adapts-guard-after-hit 4.3` The guard-change animation plays from the
      height it had, no snap or wrong transition.
- [ ] `bot-adapts-guard-after-hit 4.4` Your own guard follows the mouse at every
      height, including on the tick a hit lands.
- [ ] `bot-adapts-guard-after-hit 4.6` Judge the one-in-three; record any change
      to `GuardAdaptOdds` in design.md.
- [ ] `shove-crowded-fighters 6.2` A bot walking into you, front and back,
      standing and running: you turn to it, both arms go out, it is thrown to
      arm's length and the fight carries on face to face.
- [ ] `shove-crowded-fighters 6.3` Flush against a wall with a bot crowding in:
      nobody locks into repeated shoves.
- [ ] `shove-crowded-fighters 7.4` The shoved bot reads as stunned, not hit, and
      the shove plays to its end.
- [ ] `bot-sight-and-pathing 7.3` `go run ./cmd -level level_bot_test.json`: the
      bot ignores you until you enter its cone, follows to where it last saw you
      and gives up, jumps the wall and the gap, stops at the one it can't cross.
- [ ] `bot-sight-and-pathing 7.4` Tune `SightRange`, `SightSlope`,
      `SightSearchTicks`, `SightLingerTicks` until the bot reads as alert.

### Editor
- [ ] `main-menu-is-a-terminal 6.3` The level editor's widgets are in the dark
      style, not the menu's reds.
- [ ] `horizontal-collision 4.3` Give a tile a rectangle taller than a step on a
      floor: you run into it and stop, jump against it and climb it, walk up a
      4 px step, and `level_1.json` plays as before. Note how a fighter held
      against the wall looks.
- [ ] `reuse-object-rects 3.2` The panel shows "follows texture" / "own
      rectangles (F: follow texture)" through a reshape, then F.
- [ ] `reuse-object-rects 4.2` Ctrl+C while editing the light rectangle leaves it
      active; copy and paste work with the picker open.
- [ ] `reuse-object-rects 4.3` "Copied: <name>" appears after a copy.
- [ ] `reuse-object-rects 5.3` Shape `mn_wall1`, copy, paste onto the other seven
      `mn_*` from the open picker, place tiles, reshape `mn_wall1`: its tiles
      follow and the others don't.
- [ ] `reuse-object-rects 5.4` Reshape a tile, change its texture: it keeps its
      shape. F takes the texture's. Save, restart: a hand-edited tile is still
      own.

## Session 2: networked

Run `go run ./cmd -dedicated` plus a client, with `F5`-`F8` latency simulation
on; a third instance as spectator where noted. Every check below also means no
correction spike at the moment in question.

- [ ] `add-run-jump-and-fall 6.6` A remote run-jump and fall look the same on the
      client as on the host.
- [ ] `add-run-jump-and-fall 7.5` A remote fighter falling no longer freezes the
      other fighters on the client.
- [ ] `escalating-hit-damage 5.3` The bar drops the same on both screens as a run
      escalates; an older client is refused at the handshake.
- [ ] `fix-hitstop-input-loss 5.6` A click during hit-stop gives one attack on
      both screens; an older client is refused at the handshake.
- [ ] `flinch-without-hit-animation 4.3` Remote flinches and stuns look as they
      do on the host.
- [ ] `stun-throws-back 4.5` A stunned fighter is thrown the same distance on
      both screens.
- [ ] `bot-adapts-guard-after-hit 4.5` A bot's adapted guard looks the same on
      the client as on the host.
- [ ] `shove-crowded-fighters 6.4` Watch for a predicted turn the server does not
      make; if it shows, gate the crowding check on `W.Authority`, record it in
      design.md, and check the `A client moves nobody` scenario.
- [ ] `outline-attack-target 5.4` Each player sees the other outlined and never
      itself; a spectator sees no outlines.
- [ ] `add-spectator-mode 4.2` As spectator: pan, drag, zoom, Q/E, Space and
      follow-on-leave behave as the spec says.
- [ ] `add-spectator-mode 5.1` The spectator HUD shows with debug on and off and
      updates with the mode.
- [ ] `add-spectator-mode 5.2` Joining a listen host as spectator works; max
      spectators 0 shows "spectators full".
- [ ] `add-spectator-mode 6.1` Listen host + 1 bot, a fighter and a spectator:
      check every spectator scenario; 60 s idle keeps the spectator connected;
      quitting the host shows "server closed". Note pan speed / drag clamp
      tuning in design.md.
