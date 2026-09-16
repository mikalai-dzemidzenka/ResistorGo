## Why

The attack crosshair now pales as energy drains, but a gain is easy to miss: a correct block pays back 15 points, a shade of colour the player is not watching for in the middle of an exchange, and a pool coming back to full has no moment at all. Both are rewards the player should feel — a read that paid off, and being ready to fight again — and the crosshair is where their eyes already are.

## What Changes

- **A block that pays energy back flashes the crosshair.** When the local fighter's energy jumps up by a block's payback, a short, small ring flashes out from the attack crosshair.
- **A full pool flashes it bigger.** When the local fighter's energy comes back to full, whether by refilling or by a block, a larger, brighter ring flashes out from the crosshair. When a block both pays back and fills the pool, the full flash is the one shown.
- **Only gains the player earned flash.** Steady refilling does not flash on the way up; a respawn, which starts the new fighter full, does not flash; nor does anything shown while time is being rewound.
- The flash is drawn with the crosshair, over the finished frame, and only for the local player's own fighter.

Not in scope: flashes for other fighters, sounds, and any change to how energy is spent, restored or refilled (the refill rate is `speed-up-energy-regen`'s).

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `energy`: a new requirement that the attack crosshair flashes when the local fighter gains energy from a block and when its pool comes back to full. Added beside the crosshair requirements `tired-attack-cursor` introduced, rather than modifying any existing requirement.

## Impact

- `cursor.go`: the cursor remembers the energy it was last given and runs a flash timer — which flash, and how far through it — started when the energy it is given jumps by a block's payback or reaches full; `DrawGUI` draws the ring.
- `game.go`: `runLevel` also tells the cursor whether the energy it passes belongs to the same fighter as last frame and is live rather than a rewound frame, so a respawn or a rewind resyncs without flashing.
- Online, a block's payback reaches a client only in a snapshot; the cursor watches the energy it is shown rather than a simulation event, so the flash appears the same way online and offline. Nothing on the wire changes.
