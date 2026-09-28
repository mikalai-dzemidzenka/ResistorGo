## 1. The closed place

- [x] 1.1 Give the fighter a closed side and a countdown (2 seconds in ticks, as a named constant beside the other engagement numbers), run the countdown down every tick, and verify with a unit test that it reaches zero after exactly 120 ticks and not before
- [x] 1.2 Start the countdown when a roll lands, recording the side the roll was travelling toward, and verify with tests that a landed roll starts it on the right side while a roll interrupted by a stun or ending on a ledge does not
- [x] 1.3 Verify with a test that turning the fighter round inside the window leaves the closed side unchanged

## 2. Bots respect it

- [x] 2.1 In place claiming, refuse the closed place, refuse the far place while either is closed, and make a holder of the closed place (or of the open place from the closed side) give it up and wait in the line; verify with unit tests for each of the three
- [x] 2.2 Verify with a test that a bot approaching from the closed side waits in the line even with the open place free
- [x] 2.3 Verify with a test that once the window runs out a bot crosses to the far place again when its own is taken

## 3. The whole escape

- [x] 3.1 Add a scenario test: a fighter held at from both sides rolls out over one bot and lands; for the next 2 seconds no bot holds the place on the far side and at most one bot swings at it, and after that both places are used again. Run it repeatedly, since bots guard at random
- [ ] 3.2 Run the full test suite and the course-following bot test repeatedly and verify nothing regressed, then build with build.ps1 and check the escape by eye in play
