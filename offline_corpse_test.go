package resistor

import "testing"

// In single player a fallen bot stays on the floor for the rest of the match:
// it never flickers, never leaves the world, and nothing comes back in its place.
func TestOfflineBotStaysDead(t *testing.T) {
	offlineMatch(t)
	CL.BotsStayDead = true
	bot := CL.Spawn(CL.spawnPos(), false)
	bot.enterDeathState(nil)

	for tick := range RespawnTicks + DeathFlickerTicks + 200 {
		ticks(1)
		if bot.deathHidden() {
			t.Fatalf("the body flickered out on tick %d", tick)
		}
	}
	if CL.ByID(bot.ID) != bot {
		t.Fatalf("the body left the world")
	}
	if bot.State != StateDeath {
		t.Errorf("the body is in state %d, want death", bot.State)
	}
	if len(CL.Actors) != 2 {
		t.Errorf("%d actors in the world, want the player and the body: a bot came back", len(CL.Actors))
	}
}

// The player still falls, leaves and comes back: only bots stay down.
func TestOfflinePlayerStillRespawns(t *testing.T) {
	pl := offlineMatch(t)
	CL.BotsStayDead = true
	pl.enterDeathState(nil)

	ticks(RespawnTicks + DeathFlickerTicks + 100)
	if CL.ByID(pl.ID) != nil {
		t.Fatalf("the dead player is still in the world")
	}
	if CL.Local == nil || CL.Local == pl {
		t.Fatalf("the player did not come back")
	}
}

// A body on the floor does not hold a spawn column.
func TestSpawnIgnoresBodies(t *testing.T) {
	offlineMatch(t)
	CL.Actors[0].enterDeathState(nil)
	if p := CL.spawnPos(); p.X != 0 {
		t.Errorf("spawned at %v, want the first column the body lies on", p.X)
	}
}

// A new single-player match, and so a restart, puts every bot back on its feet.
func TestOfflineWorldStartsWithLiveBots(t *testing.T) {
	floor()
	w := newOfflineWorld()
	if !w.BotsStayDead {
		t.Errorf("single player does not keep its bodies")
	}
	if w.Local == nil || !w.Local.IsPlayer {
		t.Fatalf("no local player")
	}
	bots := 0
	for _, a := range w.Actors {
		if !a.IsPlayer {
			bots++
			if a.State == StateDeath {
				t.Errorf("bot %d starts dead", a.ID)
			}
		}
	}
	if bots != OfflineBots {
		t.Errorf("%d bots, want %d", bots, OfflineBots)
	}
}
