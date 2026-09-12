# multiplayer/interpolation Specification

## Purpose

Displays other fighters (remote players and bots) smoothly despite snapshots arriving only 20 times per second over a jittery, lossy network. Each fighter is rendered slightly in the past between known server states.

## Requirements

### Requirement: Rendering remote fighters in the past
The client SHALL render every fighter other than its own at an interpolation time. That time is the estimated current server time minus an interpolation delay: 100 ms by default, never less than two snapshot intervals. The estimate SHALL advance smoothly with real time and SHALL NOT jump backwards when snapshots arrive with jitter.

#### Scenario: Steady stream
- **WHEN** snapshots arrive at 20 Hz with ±20 ms jitter and no loss
- **THEN** remote fighters move without stutter, and their displayed state lags the server by the interpolation delay

### Requirement: Interpolated position and animation
For each remote fighter, the client SHALL find the two buffered snapshots that bracket the interpolation time. It SHALL interpolate the position linearly between them and round it to whole pixels for drawing. Discrete values SHALL come from the older of the two snapshots: state, facing, guard height and current animation. The animation frame SHALL be derived from the interpolated animation progress, so animations play at their full frame rate rather than at the snapshot rate.

#### Scenario: Walking opponent
- **WHEN** an opponent walks at constant speed
- **THEN** it is drawn moving every frame on integer pixel positions, and its walk animation shows every frame of the strip

#### Scenario: Attack animation plays fully
- **WHEN** an opponent performs an 8-frame attack lasting 0.3 s
- **THEN** all 8 frames are displayed on the observing client, not only the frames that coincide with snapshots

### Requirement: Bounded extrapolation
If no snapshot newer than the interpolation time is available, the client SHALL extrapolate a remote fighter's position from its last known velocity for at most 250 ms, then hold it in place until new data arrives. When new data arrives, the fighter SHALL return to normal interpolation without a visible jump larger than the error accumulated during extrapolation.

#### Scenario: Short loss
- **WHEN** snapshots stop for 150 ms while an opponent runs
- **THEN** the opponent continues moving in its direction during the gap

#### Scenario: Long loss
- **WHEN** snapshots stop for 2 seconds
- **THEN** remote fighters stop moving after 250 ms and stay still until snapshots resume

### Requirement: Spawns, removals and teleports
A fighter that first appears in a snapshot SHALL be shown at its spawn position once the interpolation time reaches that snapshot, without sliding from any other position. A fighter missing from snapshots SHALL disappear once the interpolation time passes its last snapshot. A position change between consecutive snapshots larger than a fighter could travel SHALL be displayed as an instant jump, not interpolated.

#### Scenario: New player appears
- **WHEN** a new player joins
- **THEN** other clients show its fighter appearing at its spawn position, not moving in from the world origin

#### Scenario: Respawn
- **WHEN** a fighter is moved by the server across the level in one tick
- **THEN** observers see it jump to the new place, not glide across the level

### Requirement: Local fighter excluded
The client's own fighter SHALL be drawn from its predicted state (see prediction), never from interpolated snapshots. It SHALL keep the existing smoothing between fixed simulation ticks for rendering.

#### Scenario: Own fighter is not delayed
- **WHEN** the player moves under 100 ms latency
- **THEN** the player's own fighter shows no interpolation delay, while opponents are shown about 100 ms behind the server
