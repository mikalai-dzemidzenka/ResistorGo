package resistor

import (
	"fmt"
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Online, the client's prediction and the server collide with the same
// rectangles. Over a floor whose every tile has two collision rectangles, its
// halves, the client's own fighter lands and stands without the server once
// correcting it.
func TestNetTwoCollisionRectFloor(t *testing.T) {
	initNet()
	defer shutdownNet()
	Stage = StageLevel // reconcile must not enter the level (needs a window)
	defer func() { Stage = StageMainMenu }()
	defer floor()

	LE = NewLevelEditor(39, 83)
	LE.objects[0] = make(LevelLayer)
	tilesPlaced := 0
	for x := int32(floorLo); x <= floorHi; x += 48 {
		c := Vector2i{X: x, Y: floorY}
		LE.objects[0][c] = LevelObject{X: c.X, Y: c.Y, CollisionRects: []rl.Rectangle{
			{Width: 24, Height: 48}, {X: 24, Width: 24, Height: 48},
		}}
		tilesPlaced++
	}
	LE.buildIndex()
	if len(LE.solids) != 2*tilesPlaced {
		t.Fatalf("%d collision rectangles for %d two-rectangle tiles", len(LE.solids), tilesPlaced)
	}

	cfg := DefaultServerConfig()
	cfg.Port = 27198
	cfg.MaxPlayers = 2
	// No bot: this is about the floor, and a bot spawns a swing away from the
	// player. Every blow it lands is a correction the client never predicted,
	// which would be counted here as the floor's doing.
	cfg.Bots = 0
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := startClient(fmt.Sprintf("127.0.0.1:%d", cfg.Port), false); err != nil {
		t.Fatal(err)
	}
	c := Cli
	defer func() { c.Close(); Cli = nil }()

	pumpFor(1500*time.Millisecond, s, c)
	if !c.Welcomed || c.W.Local == nil {
		t.Fatalf("client not in the match: welcomed=%t local=%v", c.Welcomed, c.W.Local)
	}
	if !c.W.Local.OnGround {
		t.Errorf("the client's fighter is not standing on the two-rectangle floor: pos %v", c.W.Local.Pos)
	}
	t.Logf("client seq %d, corrections %d", c.Seq, c.Corrections)
	if c.Corrections != 0 {
		t.Errorf("the server corrected the client's prediction %d times over a two-rectangle floor", c.Corrections)
	}
}
