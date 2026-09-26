## Why

Fighting is fists only. New art has arrived for a gun: a pistol torso aimed in five directions from up to down (`pistol-top.png`), and a legs layer that is only an idle pose for now (`gun-bot-idle.png`). More guns, and walk and run legs, will follow. The game needs a way to hold a gun, aim it with the mouse and fire it, and this first version should leave room for those later sheets.

## What Changes

- A fighter SHALL be able to draw and holster a pistol with a key (Q), whenever it stands or walks. There is no pickup in the level yet.
- While the pistol is out, the fighter SHALL be drawn in two layers: the gun legs, then the pistol torso on top. The torso frame SHALL be the one of five aim directions nearest to the mouse. The fighter still faces the mouse, as it does now.
- While the pistol is out, the fighter SHALL be able to walk but not run. Its legs keep the idle frame until walk and run legs exist. Jumping, rolling, climbing, being hit and dying play their usual animations, and the pistol comes back once the fighter is standing again.
- LMB SHALL fire instead of swing. A shot is a straight line from the muzzle along the aim. It stops at the first wall or the first living non-ally fighter, up to a fixed range. A fighter it stops on SHALL take a hit exactly as a landed swing lands one: guard does not stop a bullet, and the same structure cost, flinch, stun every third hit, death and kill credit apply. Shots have a short cooldown and cost no energy.
- Every shot SHALL be shown as a tracer line from the muzzle to where it stopped, plus a brief muzzle flash: a small warm light that lights the scene for a couple of frames.
- Every shot SHALL leave puffs of smoke at the muzzle that drift up and thin out, in the manner of Animal Well. A shot that stops at a wall SHALL also raise dust where it hit.
- Gun feel: the pistol SHALL kick after each shot (and the local camera with it), show a flash of fire at the muzzle, and eject a brass casing that bounces and rests. A wall hit SHALL throw sparks and leave a bullet hole. A landed shot SHALL NOT freeze the shooter the way a landed punch does.
- Multiplayer: the gun toggle and aim SHALL travel in the input command. Whether a fighter holds the gun, its aim and a shot counter SHALL travel in the snapshot. The server decides every shot against targets rewound to the shooter's view tick, the same way it does swings. **BREAKING** (network): the protocol version goes up, so older builds cannot connect.
- Out of scope: bots never use guns. There are no pickups, ammo or reload, and no gun other than the pistol.
- Every shot SHALL be heard, with one of two gunshot sounds picked at random.

## Capabilities

### New Capabilities

- `firearms`: holding and holstering a gun, aiming and the drawing of the aim frame, firing, shot resolution against walls and fighters, the tracer and the muzzle flash.

### Modified Capabilities

- `multiplayer/replication`: the input command carries the gun toggle and the aim, and the snapshot carries the held gun, the aim and the shot counter.
- `multiplayer/lag-compensation`: shots are resolved against rewound targets, as swings are.

## Impact

- `actor.go`: the `Weapon`, `Aim`, `Shots` and `ShotCooldown` fields and the `Input.Gun`/`Input.Aim` inputs; Q latching and aim sampling in `sampleInput`; the gun branch of `StateIdle` (no run, fire in place of a swing); the gun layers in `topAnim`/`topDraw`; guard ignored against shots.
- A new `gun.go`: aim and frame math, muzzle offsets, the ray cast against solids and fighters, firing, the tracer and flash state and their drawing, with a `gun_test.go` for the parts that need no window.
- `resources.go`: `gun-bot-idle.png` and `pistol-top.png` appended to `animTable`.
- `net_protocol.go`: a new button bit and aim byte in the command, three bytes more in the actor record, and `ProtocolVersion` 13.
- `net_server.go`: `runCmd` rewinds targets for a firing tick as well as for an attack.
- `game.go`: flashes added to the lights passed to `computeLights`, and tracers drawn after the light composite.
- The rim light also traces the pistol torso, because it already draws from `topDraw`.
