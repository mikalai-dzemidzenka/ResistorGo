package resistor

import (
	"fmt"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// dropFloor installs a drop-through platform under every spawn column, its top
// at floorY, with solid floor 100 px under it.
func dropFloor() {
	LE = NewLevelEditor(39, 83)
	LE.objects[0] = make(LevelLayer)
	for x := int32(floorLo); x <= floorHi; x += 48 {
		LE.objects[0][Vector2i{X: x, Y: floorY}] = LevelObject{X: x, Y: floorY, DropRects: []rl.Rectangle{{Width: 48, Height: 16}}}
		LE.objects[0][Vector2i{X: x, Y: floorY + 100}] = LevelObject{X: x, Y: floorY + 100, CollisionRects: []rl.Rectangle{{Width: 48, Height: 48}}}
	}
	LE.buildIndex()
}

// A player presses S online: the server's fighter climbs down and falls past
// the platform, the player's own prediction goes with it rather than being put
// back on top, and another player is sent the climb-down to draw.
func TestNetClimbDown(t *testing.T) {
	initNet()
	defer shutdownNet()
	Stage = StageLevel
	defer func() { Stage = StageMainMenu }()
	dropFloor()
	defer floor()

	cfg := DefaultServerConfig()
	cfg.Port = 27197
	cfg.MaxPlayers = 2
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { Srv = nil; s.Close() }()
	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port), false); err != nil {
		t.Fatal(err)
	}
	c := Cli
	pumpFor(1500*time.Millisecond, s, c)
	if !c.Welcomed || c.W.Local == nil {
		t.Fatal("client not in the match")
	}
	var srvA *Actor
	for _, p := range s.Peers {
		if p.Actor != nil {
			srvA = p.Actor
		}
	}
	if srvA == nil || feet(srvA) != floorY || feet(c.W.Local) != floorY {
		t.Fatalf("not standing on the platform: server %v, client %v", srvA, feet(c.W.Local))
	}
	other := hello(t, s, c, cfg.Port, Hello{Proto: ProtocolVersion, LevelHash: LE.Hash()})
	defer other.h.Destroy()
	pumpFor(300*time.Millisecond, s, c, other)
	other.got = nil

	corr := c.Corrections
	pressedDown = true
	climbed, below := false, false
	for range 40 {
		pumpFor(50*time.Millisecond, s, c, other)
		climbed = climbed || srvA.State == StateClimbDown
		if f := feet(c.W.Local); f > floorY {
			below = true
		} else if below {
			t.Fatalf("the client's fighter was put back on the platform: feet %v", f)
		}
	}
	if !climbed {
		t.Error("the server's fighter never climbed down")
	}
	if feet(srvA) != floorY+100 || feet(c.W.Local) != floorY+100 {
		t.Errorf("after the climb-down: server feet %v, client feet %v; want both on the floor at %v", feet(srvA), feet(c.W.Local), floorY+100)
	}
	t.Logf("corrections during the climb-down: %d", c.Corrections-corr)

	seen := false
	id := animID("resources/threefour/climb-down.png")
	for _, b := range other.got {
		if len(b) == 0 || b[0] != MsgSnapshot {
			continue
		}
		snap, err := decodeSnapshot(b)
		if err != nil {
			t.Fatalf("a snapshot the other player got is malformed: %v", err)
		}
		for _, r := range snap.Actors {
			if r.ID == srvA.ID && r.Anim == id && ActorState(r.State) == StateClimbDown {
				seen = true
			}
		}
	}
	if !seen {
		t.Error("the other player was never sent the climb-down")
	}
}
