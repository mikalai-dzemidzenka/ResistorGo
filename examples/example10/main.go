package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"time"

	enet "github.com/codecat/go-enet"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	port       = 7777
	tickRate   = 64
	tickPeriod = time.Second / tickRate
	tickSecs   = 1.0 / tickRate
	renderLag  = 0.10
)

type Player struct {
	ID   uint32
	X, Y float32
}

type Snapshot struct {
	Tick    uint64
	Players []Player
}

type Client struct {
	Player Player
	Seen   time.Time
	Send   func([]byte)
}

func pack(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

func main() {
	server := flag.Bool("server", false, "Запустить сервер")
	ip := flag.String("ip", "127.0.0.1", "Адрес сервера")
	flag.Parse()

	enet.Initialize()
	defer enet.Deinitialize()

	if *server {
		serve()
	} else {
		play(*ip)
	}
}

func serve() {
	host, err := enet.NewHost(
		enet.NewListenAddress(port), 32, 1, 0, 0,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer host.Destroy()

	clients := map[uint32]Client{}
	start := time.Now()
	ticker := time.NewTicker(tickPeriod)
	defer ticker.Stop()

	log.Printf("UDP :%d", port)

	for {
		// Ограничиваем обработку сети, чтобы не задерживать тики.
		for i := 0; i < 64; i++ {
			event := host.Service(1)
			if event.GetType() == enet.EventNone {
				break
			}
			if event.GetType() != enet.EventReceive {
				continue
			}

			packet := event.GetPacket()
			var p Player
			err := json.Unmarshal(packet.GetData(), &p)
			packet.Destroy()
			if err != nil {
				continue
			}

			peer := event.GetPeer()
			clients[p.ID] = Client{
				Player: p,
				Seen:   time.Now(),
				Send: func(data []byte) {
					peer.SendBytes(data, 0, 0)
				},
			}
		}

		select {
		case stamp := <-ticker.C:
			// Пропущенные при перегрузке тики не растягивают время.
			snapshot := Snapshot{
				Tick: uint64(stamp.Sub(start) / tickPeriod),
			}

			for id, client := range clients {
				if time.Since(client.Seen) > 3*time.Second {
					delete(clients, id)
					continue
				}
				snapshot.Players = append(
					snapshot.Players, client.Player,
				)
			}

			data := pack(snapshot)
			for _, client := range clients {
				client.Send(data)
			}

		default:
		}
		// Очередь отправки обслужит следующий host.Service.
	}
}

// Интерполяция по серверному времени, не по времени прихода пакетов.
func interpolate(buf []Snapshot, target float64) []Player {
	if len(buf) == 0 {
		return nil
	}

	if target <= float64(buf[0].Tick)*tickSecs {
		return buf[0].Players
	}

	for i := 1; i < len(buf); i++ {
		a, b := buf[i-1], buf[i]
		bt := float64(b.Tick) * tickSecs
		if target > bt {
			continue
		}

		at := float64(a.Tick) * tickSecs
		t := float32((target - at) / (bt - at))

		previous := make(map[uint32]Player, len(a.Players))
		for _, p := range a.Players {
			previous[p.ID] = p
		}

		result := make([]Player, 0, len(b.Players))
		for _, p := range b.Players {
			if old, exists := previous[p.ID]; exists {
				p.X = old.X + (p.X-old.X)*t
				p.Y = old.Y + (p.Y-old.Y)*t
			}
			result = append(result, p)
		}
		return result
	}

	// Буфер закончился: удерживаем последнюю позицию.
	return buf[len(buf)-1].Players
}

func play(ip string) {
	host, err := enet.NewHost(nil, 1, 1, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer host.Destroy()

	peer, err := host.Connect(enet.NewAddress(ip, port), 1, 0)
	if err != nil {
		log.Fatal(err)
	}

	rl.InitWindow(800, 450, "Network platformer")
	defer rl.CloseWindow()
	rl.SetTargetFPS(128)

	p := Player{ID: rand.Uint32(), X: 100, Y: 100}
	var vy float32
	var sendTimer float64
	grounded, connected := false, false

	var snapshots []Snapshot
	var clockOffset float64
	clockReady := false

	platforms := []rl.Rectangle{
		{X: 0, Y: 410, Width: 800, Height: 40},
		{X: 260, Y: 310, Width: 180, Height: 20},
		{X: 520, Y: 220, Width: 160, Height: 20},
	}

	for !rl.WindowShouldClose() {
		for i := 0; i < 64; i++ {
			event := host.Service(0)
			if event.GetType() == enet.EventNone {
				break
			}

			switch event.GetType() {
			case enet.EventConnect:
				connected = true

			case enet.EventDisconnect:
				connected = false
				snapshots = nil
				clockReady = false

			case enet.EventReceive:
				packet := event.GetPacket()
				var snapshot Snapshot
				err := json.Unmarshal(packet.GetData(), &snapshot)
				packet.Destroy()
				if err != nil {
					continue
				}

				n := len(snapshots)
				if n > 0 && snapshot.Tick <= snapshots[n-1].Tick {
					continue
				}

				// Минимальная наблюдаемая разница часов:
				// задержанные пакеты не сдвигают шкалу времени назад.
				offset := rl.GetTime() -
					float64(snapshot.Tick)*tickSecs
				if !clockReady || offset < clockOffset {
					clockOffset = offset
					clockReady = true
				}

				snapshots = append(snapshots, snapshot)
				if len(snapshots) > 32 {
					snapshots = snapshots[len(snapshots)-32:]
				}
			}
		}

		frameDt := rl.GetFrameTime()
		dt := frameDt
		if dt > 0.033 {
			dt = 0.033
		}

		// Локальное движение — сразу, без сетевой задержки.
		if rl.IsKeyDown(rl.KeyA) {
			p.X -= 230 * dt
		}
		if rl.IsKeyDown(rl.KeyD) {
			p.X += 230 * dt
		}
		if p.X < 0 {
			p.X = 0
		}
		if p.X > 776 {
			p.X = 776
		}

		if grounded && rl.IsKeyPressed(rl.KeySpace) {
			vy = -480
		}

		oldBottom := p.Y + 32
		vy += 1200 * dt
		p.Y += vy * dt
		grounded = false

		// Односторонние платформы: столкновение только сверху.
		for _, r := range platforms {
			if vy >= 0 &&
				p.X+24 > r.X && p.X < r.X+r.Width &&
				oldBottom <= r.Y && p.Y+32 >= r.Y {
				p.Y = r.Y - 32
				vy = 0
				grounded = true
			}
		}

		sendTimer += float64(frameDt)
		if connected && sendTimer >= tickSecs {
			sendTimer -= tickSecs
			peer.SendBytes(pack(p), 0, 0)
		}

		targetTime := rl.GetTime() - clockOffset - renderLag
		others := interpolate(snapshots, targetTime)

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		for _, r := range platforms {
			rl.DrawRectangleRec(r, rl.DarkGray)
		}

		for _, other := range others {
			if other.ID == p.ID {
				continue
			}
			// Float-координаты сохраняют плавность движения.
			rl.DrawRectangleV(
				rl.NewVector2(other.X, other.Y),
				rl.NewVector2(24, 32),
				rl.Orange,
			)
		}

		rl.DrawRectangleV(
			rl.NewVector2(p.X, p.Y),
			rl.NewVector2(24, 32),
			rl.Green,
		)

		rl.DrawText("A/D: move | SPACE: jump", 10, 10, 20, rl.DarkGray)
		rl.DrawText("20 Hz | interpolation 100 ms", 10, 35, 20, rl.DarkGray)

		if !connected {
			rl.DrawText("Offline / connecting", 10, 60, 20, rl.Red)
		}

		rl.EndDrawing()
	}
}
