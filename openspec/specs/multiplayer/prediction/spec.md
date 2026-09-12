# multiplayer/prediction Specification

## Purpose

Keeps the local fighter responsive under latency. The client simulates its own inputs immediately and then reconciles with the authoritative server state without visible jitter.

## Requirements

### Requirement: Immediate local response
A networked client SHALL apply its own input to its fighter on the same client tick the input is sampled, without waiting for the server. This covers movement, jumping, guard height changes, and starting and chaining attacks.

#### Scenario: Movement under latency
- **WHEN** artificial latency is 200 ms and the player presses D
- **THEN** the local fighter starts moving on the next simulation tick

#### Scenario: Attack under latency
- **WHEN** artificial latency is 200 ms and the player clicks attack
- **THEN** the local fighter's attack animation starts on the next simulation tick

### Requirement: Command history
The client SHALL keep every command it has sent but that the server has not yet acknowledged, for at least 1 second of ticks. If the history overflows, the oldest entries SHALL be discarded.

#### Scenario: Acknowledged commands are dropped
- **WHEN** a snapshot acknowledges command sequence N
- **THEN** commands with sequence ≤ N are removed from the history

### Requirement: Reconciliation
On receiving a snapshot, the client SHALL reset its fighter to the snapshot's authoritative state and re-simulate all unacknowledged commands in order. The reconciled result SHALL become the current predicted state. If prediction was correct, the reconciled state SHALL equal the pre-correction predicted state, and no visible change SHALL occur.

#### Scenario: Correct prediction
- **WHEN** the player walks on flat ground with no other fighters nearby, under 150 ms latency
- **THEN** no prediction corrections are counted and the fighter moves smoothly

#### Scenario: Hit by another player
- **WHEN** the server puts the local fighter into hit-stun because of an opponent's attack the client did not predict
- **THEN** after the snapshot arrives the local fighter shows hit-stun, with any unacknowledged inputs re-applied on top of that state

#### Scenario: Blocked attack
- **WHEN** the local fighter's predicted attack is blocked on the server
- **THEN** the fighter's state converges to the server's outcome after reconciliation

### Requirement: Correction smoothing
When reconciliation changes the local fighter's displayed position by at most 8 pixels, the client SHALL blend the difference out over about 100 ms. Larger differences SHALL be applied immediately. The simulation state SHALL always be the corrected one; only the drawn position is blended.

#### Scenario: Small correction
- **WHEN** reconciliation moves the fighter by 3 pixels
- **THEN** the drawn fighter slides to the corrected position over about 100 ms instead of jumping

#### Scenario: Large correction
- **WHEN** reconciliation moves the fighter by 30 pixels
- **THEN** the drawn fighter is placed at the corrected position immediately

### Requirement: Server-confirmed combat outcomes
The client SHALL NOT decide hits or blocks on other fighters. Hit reactions SHALL appear only after the server's outcome is received: hit flash, hit-stop, camera shake, knockback, and block animations on opponents. Re-simulation during reconciliation SHALL NOT re-trigger presentation effects (camera zoom, camera shake, sounds) that already played.

#### Scenario: No false hit
- **WHEN** the local player swings at an opponent who, on the server, has moved out of range
- **THEN** the opponent never shows a hit reaction on the attacker's screen

#### Scenario: Replay does not repeat effects
- **WHEN** reconciliation replays 10 commands that include an attack start
- **THEN** the attack zoom effect is not triggered a second time
