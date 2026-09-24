package resistor

import (
	"slices"

	lua "github.com/Shopify/go-lua"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// registerScriptAPI gives a script the functions it reads and changes the
// match with. Every id a script is given can go stale - a fighter dies, a
// player leaves - so every function takes a stale id quietly and answers nil
// or false, rather than making the script check first.
//
// A function writes the same fields the game does, and the snapshots carry
// them to the clients as they carry any other change.
func registerScriptAPI(s *script) {
	h := s.host
	l := s.l
	fn := func(name string, f lua.Function) { l.Register(name, f) }

	fn("GetTick", func(l *lua.State) int {
		l.PushInteger(int(h.tick()))
		return 1
	})

	// players
	fn("GetPlayers", func(l *lua.State) int {
		pushIDs(l, h.playerIDs())
		return 1
	})
	fn("IsPlayerConnected", func(l *lua.State) int {
		l.PushBoolean(h.player(l, 1) != nil)
		return 1
	})
	fn("GetPlayerFighter", func(l *lua.State) int {
		if p := h.player(l, 1); p != nil && p.Actor != nil {
			l.PushInteger(int(p.Actor.ID))
		} else {
			l.PushNil()
		}
		return 1
	})
	fn("SendPlayerMessage", func(l *lua.State) int {
		p := h.player(l, 1)
		text := lua.CheckString(l, 2)
		if p != nil {
			h.srv.sendText(p, text)
		}
		l.PushBoolean(p != nil)
		return 1
	})
	fn("SendMessageToAll", func(l *lua.State) int {
		text := lua.CheckString(l, 1)
		if h.srv != nil {
			h.srv.sendText(nil, text)
		}
		return 0
	})
	fn("KickPlayer", func(l *lua.State) int {
		l.PushBoolean(h.kick(l, 1))
		return 1
	})

	// fighters
	fn("GetFighters", func(l *lua.State) int {
		ids := make([]uint16, 0, len(h.w.Actors))
		for _, a := range h.w.Actors {
			ids = append(ids, a.ID)
		}
		pushIDs(l, ids)
		return 1
	})
	fn("IsBot", func(l *lua.State) int {
		a := h.fighter(l, 1)
		l.PushBoolean(a != nil && !a.IsPlayer)
		return 1
	})
	fn("GetFighterPlayer", func(l *lua.State) int {
		if a := h.fighter(l, 1); a != nil {
			l.PushInteger(int(a.Owner))
		} else {
			l.PushNil()
		}
		return 1
	})
	fn("GetFighterPos", func(l *lua.State) int {
		a := h.fighter(l, 1)
		if a == nil {
			l.PushNil()
			return 1
		}
		l.PushNumber(float64(a.Pos.X))
		l.PushNumber(float64(a.Pos.Y))
		return 2
	})
	fn("SetFighterPos", func(l *lua.State) int {
		a := h.living(l, 1)
		x, y := lua.CheckNumber(l, 2), lua.CheckNumber(l, 3)
		l.PushBoolean(a != nil && placeFighter(a, rl.Vector2{X: float32(x), Y: float32(y)}))
		return 1
	})
	fn("GetFighterStructure", func(l *lua.State) int {
		pushFighterNumber(l, h.fighter(l, 1), func(a *Actor) float64 { return float64(a.Structure) })
		return 1
	})
	fn("SetFighterStructure", func(l *lua.State) int {
		a := h.living(l, 1)
		v := float32(lua.CheckNumber(l, 2))
		if a == nil {
			l.PushBoolean(false)
			return 1
		}
		// Nothing standing has no structure: a fighter set to none is broken,
		// as KillFighter breaks it.
		if v <= 0 {
			a.Structure = 0
			a.enterDeathState(nil)
		} else {
			a.Structure = min(v, MaxStructure)
		}
		l.PushBoolean(true)
		return 1
	})
	fn("GetFighterEnergy", func(l *lua.State) int {
		pushFighterNumber(l, h.fighter(l, 1), func(a *Actor) float64 { return float64(a.Energy) })
		return 1
	})
	fn("SetFighterEnergy", func(l *lua.State) int {
		a := h.living(l, 1)
		v := lua.CheckNumber(l, 2)
		if a != nil {
			a.Energy = int32(clamp(v, 0, MaxEnergy))
		}
		l.PushBoolean(a != nil)
		return 1
	})
	fn("GetFighterTeam", func(l *lua.State) int {
		pushFighterNumber(l, h.fighter(l, 1), func(a *Actor) float64 { return float64(a.Team) })
		return 1
	})
	fn("SetFighterTeam", func(l *lua.State) int {
		a := h.fighter(l, 1)
		team := lua.CheckInteger(l, 2)
		if a != nil {
			a.Team = uint8(clamp(team, 0, 255))
		}
		l.PushBoolean(a != nil)
		return 1
	})
	fn("KillFighter", func(l *lua.State) int {
		a := h.living(l, 1)
		if a != nil {
			a.Structure = 0
			a.enterDeathState(nil)
		}
		l.PushBoolean(a != nil)
		return 1
	})
	fn("AddBot", func(l *lua.State) int {
		pos := h.w.spawnPos()
		if !l.IsNoneOrNil(1) {
			pos = rl.Vector2{X: float32(lua.CheckNumber(l, 1)), Y: float32(lua.OptNumber(l, 2, 0))}
		}
		a := h.w.Spawn(pos, false)
		if h.srv != nil {
			h.srv.Bots = append(h.srv.Bots, a)
		}
		l.PushInteger(int(a.ID))
		return 1
	})
	fn("RemoveBot", func(l *lua.State) int {
		a := h.living(l, 1)
		if a == nil || a.IsPlayer {
			l.PushBoolean(false)
			return 1
		}
		h.w.Remove(a)
		if h.srv != nil {
			h.srv.Bots = slices.DeleteFunc(h.srv.Bots, func(b *Actor) bool { return b == a })
		}
		l.PushBoolean(true)
		return 1
	})
}

// fighter is the fighter the id at arg names, or nil.
func (h *scriptHost) fighter(l *lua.State, arg int) *Actor {
	id := lua.CheckInteger(l, arg)
	if id <= 0 || id > 0xffff {
		return nil
	}
	return h.w.ByID(uint16(id))
}

// living is the fighter at arg if it is still standing: one already broken is
// only waiting to be taken away, and there is nothing left to change about it.
func (h *scriptHost) living(l *lua.State, arg int) *Actor {
	if a := h.fighter(l, arg); a != nil && a.State != StateDeath {
		return a
	}
	return nil
}

// player is the connected fighter player the id at arg names, or nil.
func (h *scriptHost) player(l *lua.State, arg int) *netPlayer {
	id := lua.CheckInteger(l, arg)
	if h.srv == nil {
		return nil
	}
	for _, p := range h.srv.Peers {
		if p.Accepted && p.Role == RoleFighter && int(p.ID) == id {
			return p
		}
	}
	return nil
}

func (h *scriptHost) playerIDs() []uint16 {
	var ids []uint16
	if h.srv != nil {
		for _, p := range h.srv.Peers {
			if p.Accepted && p.Role == RoleFighter {
				ids = append(ids, p.ID)
			}
		}
	}
	slices.Sort(ids)
	return ids
}

// kick drops the player at arg: it is told the connection is over, and the
// scripts hear it leave with the reason "kicked".
func (h *scriptHost) kick(l *lua.State, arg int) bool {
	p := h.player(l, arg)
	if p == nil {
		return false
	}
	for key, q := range h.srv.Peers {
		if q == p {
			if p.Peer != nil {
				p.Peer.DisconnectLater(0)
			}
			h.srv.drop(key, "kicked")
			break
		}
	}
	return true
}

// placeFighter moves a to pos, standing still, unless its body there would be
// inside the level's solid ground: a fighter put into a wall would be stuck in
// it, and nothing in the game would ever get it out.
func placeFighter(a *Actor, pos rl.Vector2) bool {
	was := a.Pos
	a.Pos = pos
	body := a.collisionRect()
	for _, c := range LE.solids {
		if aabb(body, c) {
			a.Pos = was
			return false
		}
	}
	a.PrevPos, a.InterpolatedPos = pos, pos
	a.Speed = rl.Vector2{}
	a.checkCollision()
	return true
}

func pushIDs(l *lua.State, ids []uint16) {
	l.CreateTable(len(ids), 0)
	for i, id := range ids {
		l.PushInteger(int(id))
		l.RawSetInt(-2, i+1)
	}
}

func pushFighterNumber(l *lua.State, a *Actor, get func(*Actor) float64) {
	if a == nil {
		l.PushNil()
		return
	}
	l.PushNumber(get(a))
}
