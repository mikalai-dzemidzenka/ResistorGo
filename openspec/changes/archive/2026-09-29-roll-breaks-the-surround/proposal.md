## Why

A roll out of a crowd buys distance and nothing else: both places at the
fighter stay held, and the bot whose place ends up on the far side crosses
straight through to get behind it again, so the fighter is surrounded within
about a second of landing. Rolling out should be a way to break an
encirclement, not just to move it.

## What Changes

- A fighter that lands a roll closes the place on the side it rolled toward
  for 2 seconds. While it is closed, bots fight that fighter from the open
  side only.
- While a place is closed, a bot SHALL NOT take it, SHALL NOT cross round to
  it, and a bot holding it SHALL give it up and wait in the line.
- A bot standing on the closed side waits there rather than crossing through
  the fighter to the open place.
- Only a roll that lands closes a place: one interrupted by a stun or a death,
  or one that ends by catching a ledge, does not. A hit that only flinches
  leaves the roll going, and it closes the place when it lands.
- When the 2 seconds run out, both places open again and the usual rules,
  crossing included, resume.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `engagement`: a new requirement for the closed place after a landed roll,
  and the crossing rule gains the exception that it never leads to a closed
  place.
- `rolling`: landing from a roll starts the 2-second window that closes the
  place ahead of it.

## Impact

- Code: the fighter gains a closed side and a countdown, started when a roll
  lands and run down every tick; the bot place-claiming logic reads them.
- Networking: none. Places are claimed only by the authority, so the new
  state is never sent in a snapshot and the protocol does not change.
- Rewind and recording carry the new fields automatically, since a fighter is
  recorded by value.
