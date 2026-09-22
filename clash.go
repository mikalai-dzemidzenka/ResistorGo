package resistor

const (
	clashLeftAnim  = "resources/threefour/battle-low-1-left.png"
	clashRightAnim = "resources/threefour/battle-low-1-right.png"
)

// openingLowSwing reports whether the fighter is in the first blow of a low
// attack: the one enterAttackState started. A chained blow or a change of
// height replaces the animation, so the animation alone says which it is.
func (a *Actor) openingLowSwing() bool {
	return a.State == StateAttack && a.Anim == "resources/threefour/attack-low.png"
}

// canClash reports whether the fighter is free to be drawn into a clash at all.
func (a *Actor) canClash() bool {
	switch a.State {
	case StateClash, StateAttacked, StateDeath:
		return false
	}
	return a.OnGround
}

// tryClash starts a clash between a, in its opening low swing, and b, if b
// answered it in time: with an opening low swing of its own inside the window,
// or by asking for a low swing from the block a's swing put it in. It is the
// one place the conditions live, whichever update finds the pair. Only the
// authority decides a clash, as it decides hits.
func tryClash(a, b *Actor) bool {
	if a == b || a.W == nil || !a.W.Authority || a.allyOf(b) {
		return false
	}
	if !a.canClash() || !b.canClash() || !a.openingLowSwing() {
		return false
	}
	if a.Orientation != facing(b.Pos.X-a.Pos.X) || b.Orientation != facing(a.Pos.X-b.Pos.X) {
		return false
	}
	if !a.inSwingRange(b) || !b.inSwingRange(a) {
		return false
	}
	switch {
	case b.openingLowSwing():
		gap := int64(a.SwingTick) - int64(b.SwingTick)
		if gap < -ClashWindowTicks || gap > ClashWindowTicks {
			return false
		}
	case b.State == StateBlock:
		if b.Attacker != a || !b.Input.Attack || b.Updown != 0 || !b.canSwing() ||
			a.W.Tick-a.SwingTick > ClashWindowTicks {
			return false
		}
	default:
		return false
	}
	startClash(a, b)
	return true
}

// startClash pays for the clash and starts it for both fighters on the same
// tick. A fighter answering from a block never paid for its swing and pays for
// it here. Structure the clash breaks kills, the other taking the kill, and
// then there is no clash to play.
func startClash(a, b *Actor) {
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
		frames uint8
	}{
		{left, mid - ClashGap/2, 1, clashLeftAnim, ClashLeftFrames},
		{right, mid + ClashGap/2, -1, clashRightAnim, ClashRightFrames},
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
		f.SetAnim(s.anim, AnimOptions{animDuration: ClashSeconds, totalFrames: s.frames, once: true, reset: true})
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
