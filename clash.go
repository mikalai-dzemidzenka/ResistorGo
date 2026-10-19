package resistor

// clashKind is one height's clash: the sheet the fighter on the left plays and
// the one the fighter on the right plays (the two share frame count and
// timing), and impacts, the frames where a blow meets the other fighter and a
// punch is heard. Timings and impacts are to be tuned by eye in play.
type clashKind struct {
	left, right string
	frames      uint8
	seconds     float64 // 70 ms a frame
	impacts     []uint8
}

// clashKinds is indexed by the height both fighters swing at (Updown): low and
// mid clash, high does not.
var clashKinds = [...]clashKind{
	// the left fighter's punch meeting the knee, and the right fighter's two
	// kicks landing on the guard
	0: {"resources/threefour/battle-low-1-left.png", "resources/threefour/battle-low-1-right.png", 13, 0.91, []uint8{4, 7, 10}},
	// the left fighter's kick landing on the right one's guard
	1: {"resources/threefour/battle-mid-1-left.png", "resources/threefour/battle-mid-1-right.png", 9, 0.63, []uint8{4}},
}

// openingSwing reports the height of the fighter's first blow of a low or mid
// attack - the one enterAttackState started - and whether it is in one. A
// chained blow or a change of height replaces the animation, so the animation
// alone says which it is.
func (a *Actor) openingSwing() (uint8, bool) {
	if a.State != StateAttack {
		return 0, false
	}
	switch a.Anim {
	case "resources/threefour/attack-low.png":
		return 0, true
	case "resources/threefour/attack-mid.png":
		return 1, true
	}
	return 0, false
}

// clashImpacts are the impacts of the clash the fighter's animation belongs to,
// counted on the left fighter's sheet only.
func clashImpacts(anim string) []uint8 {
	for _, k := range clashKinds {
		if anim == k.left {
			return k.impacts
		}
	}
	return nil
}

// canClash reports whether the fighter is free to be drawn into a clash at all.
func (a *Actor) canClash() bool {
	switch a.State {
	case StateClash, StateAttacked, StateDeath:
		return false
	}
	return a.OnGround
}

// tryClash starts a clash between a, in its opening low or mid swing, and b,
// if b answered it in time at the same height: with an opening swing of its own
// inside the window, or by asking for a swing from the block a's swing put it
// in. It is the one place the conditions live, whichever update finds the
// pair. Only the authority decides a clash, as it decides hits.
func tryClash(a, b *Actor) bool {
	if a == b || a.W == nil || !a.W.Authority || a.allyOf(b) {
		return false
	}
	h, ok := a.openingSwing()
	if !ok || !a.canClash() || !b.canClash() {
		return false
	}
	if a.Orientation != facing(b.Pos.X-a.Pos.X) || b.Orientation != facing(a.Pos.X-b.Pos.X) {
		return false
	}
	if !a.inSwingRange(b) || !b.inSwingRange(a) {
		return false
	}
	bh, bOpening := b.openingSwing()
	switch {
	case bOpening && bh == h:
		gap := int64(a.SwingTick) - int64(b.SwingTick)
		if gap < -ClashWindowTicks || gap > ClashWindowTicks {
			return false
		}
	case b.State == StateBlock:
		if b.Attacker != a || !b.Input.Attack || b.Updown != h || !b.canSwing() ||
			a.W.Tick-a.SwingTick > ClashWindowTicks {
			return false
		}
	default:
		return false
	}
	startClash(a, b, clashKinds[h])
	return true
}

// startClash pays for the clash and starts it for both fighters on the same
// tick. A fighter answering from a block never paid for its swing and pays for
// it here. Structure the clash breaks kills, the other taking the kill, and
// then there is no clash to play.
func startClash(a, b *Actor, k clashKind) {
	if b.State == StateBlock {
		b.spendEnergy(EnergySwing)
	}
	aBroke := a.spendStructure(ClashStructure)
	bBroke := b.spendStructure(ClashStructure)
	if aBroke || bBroke {
		for _, f := range []struct {
			self, other *Actor
			broke       bool
		}{{a, b, aBroke}, {b, a, bBroke}} {
			if f.broke {
				f.self.enterDeathState(f.other)
			} else {
				f.self.leaveSwingOrBlock()
				f.self.enterIdleState()
			}
		}
		return
	}

	left, right := a, b
	if b.Pos.X < a.Pos.X {
		left, right = b, a
	}
	mid := (left.Pos.X + right.Pos.X) / 2
	for _, s := range []struct {
		f      *Actor
		x      float32
		orient int8
		anim   string
	}{
		{left, mid - ClashGap/2, 1, k.left},
		{right, mid + ClashGap/2, -1, k.right},
	} {
		f := s.f
		f.leaveSwingOrBlock()
		from := f.Pos.X
		f.Pos.X = s.x
		f.Speed.X = 0
		f.blockSides(from)
		f.Orientation = s.orient
		f.State = StateClash
		f.stopFlicker()
		f.AttackQueued, f.BufAttack, f.BufJump, f.SwingHit = false, false, false, false
		f.SetAnim(s.anim, AnimOptions{animDuration: k.seconds, totalFrames: k.frames, once: true, reset: true})
	}
	left.Attacker, right.Attacker = right, left
}

// leaveSwingOrBlock ends what the fighter was doing when the clash found it.
func (a *Actor) leaveSwingOrBlock() {
	switch a.State {
	case StateAttack:
		a.exitAttackState()
	case StateBlock:
		a.exitBlockState()
	}
}

// endClash puts both fighters back on their feet together. The two sheets are
// started on the same tick but each is timed from its own fighter's tick count,
// and the fighter updated later that tick starts one tick ahead; so whichever
// finishes first ends the clash for both, and neither is left alone in it.
func (a *Actor) endClash() {
	if o := a.Attacker; o != nil && o.State == StateClash && o.Attacker == a {
		o.Attacker = nil
		o.enterIdleState()
	}
	a.Attacker = nil
	a.enterIdleState()
}
