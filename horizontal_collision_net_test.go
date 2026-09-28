package resistor

import (
	"fmt"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Online against a wall: the client's own fighter, held running into a wall,
// stops against it at the same X on the client and on the server, and the
// server never has to correct the client's prediction.
func TestNetRunIntoAWall(t *testing.T) {
	initNet()
	defer shutdownNet()
	Stage = StageLevel // reconcile must not enter the level (needs a window)
	defer func() { Stage = StageMainMenu }()
	walls()
	defer floor()

	cfg := DefaultServerConfig()
	cfg.Port = 27200
	cfg.MaxPlayers = 2
	cfg.Bots = 0 // a bot's blows are corrections the client never predicted
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { Srv = nil; s.Close() }()
	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port), false); err != nil {
		t.Fatal(err)
	}
	c := Cli
	defer func() { c.Close(); Cli = nil }()

	pumpFor(1500*time.Millisecond, s, c)
	if !c.Welcomed || c.W.Local == nil {
		t.Fatalf("client not in the match: welcomed=%t local=%v", c.Welcomed, c.W.Local)
	}
	var srvA *Actor
	for _, p := range s.Peers {
		if p.Actor != nil {
			srvA = p.Actor
		}
	}
	if srvA == nil {
		t.Fatal("no fighter on the server")
	}
	if boxLeft(c.W.Local) <= wallX {
		t.Fatalf("spawned at or past the wall: left side %v, wall face %v", boxLeft(c.W.Local), wallX)
	}

	// KeyA moves left, towards the wall's right face. The attack cursor goes
	// far below the fighter: left at the screen's top it guards block-all,
	// which does not move.
	keyDown = func(k int32) bool { return k == rl.KeyA }
	defer func() { keyDown = rl.IsKeyDown }()
	atk := Cur.AtkPos
	Cur.AtkPos.Y = 1e5
	defer func() { Cur.AtkPos = atk }()
	pumpFor(2*time.Second, s, c)

	if boxLeft(c.W.Local) != wallX || boxLeft(srvA) != wallX {
		t.Errorf("held into the wall: client left side %v, server %v; want both at the face, %v",
			boxLeft(c.W.Local), boxLeft(srvA), wallX)
	}
	if feet(c.W.Local) != floorY || feet(srvA) != floorY {
		t.Errorf("not on the floor against the wall: client feet %v, server %v", feet(c.W.Local), feet(srvA))
	}
	t.Logf("client seq %d, corrections %d", c.Seq, c.Corrections)
	if c.Corrections != 0 {
		t.Errorf("the server corrected the client's prediction %d times running into a wall", c.Corrections)
	}
}
