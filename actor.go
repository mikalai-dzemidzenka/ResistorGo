package resistor

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"math/rand"
	"strings"
)

const (
	DefaultAcc      = 270
	DefaultFriction = 250
	RunReach        = 200
	MaxSpeed        = 320
	PunchSpeed      = 50
	AttackDash      = 4
	// AttackMoveSpeed is how fast a fighter can move while it swings, either
	// way, without turning: about 6 px over a low swing, enough to take or give
	// up the edge of a reach and not enough to walk out of a fight.
	AttackMoveSpeed = 20
	// WalkStopSpeed is where a fighter walking on its own with no key held
	// stops, rather than sliding on down through the slowest speeds.
	WalkStopSpeed = 60
	StepHeight    = 4 // a rise no taller than this is walked onto; anything taller is a wall
	// The run-jump leaves the ground on the third frame of its animation: the
	// frame the fighter pushes off on. The two before it are the anticipation,
	// drawn rather than spent in a state of its own.
	RunJumpLaunchFrame = 2
	// How long each jump is drawn winding up before it leaves the ground. Named
	// here rather than left in the two SetAnim calls because the wind-up is not
	// only an animation: it is ground the fighter covers still standing, which
	// is what a bot has to press its jump ahead of. See botJumpWindupTicks.
	RunJumpSeconds   = 0.5
	RunJumpFrames    = 6
	JumpStartSeconds = 0.7
	JumpStartFrames  = 13
	Gravity          = 1500
	MaxJumpSpeed     = -Gravity * 10
	// The run-jump leaves the ground harder than a standing one: the run is
	// what buys the extra height, and the half-pool it costs is what the
	// height is paid for with. Derived from the standing jump, so retuning
	// that carries the run-jump along.
	RunJumpLift     = 1.75
	MaxRunJumpSpeed = MaxJumpSpeed * RunJumpLift
	MaxFallSpeed    = 900
	AttackZoomSpeed = 2.0

	BackAttackSpeed   = 0.2
	SecondAttackSpeed = 0.2
	BlockSpeed        = 0.2

	HitStopTicks = 6

	// Structure is the fighter's only health model, counted in hit-units.
	// A hit costs 1, block-all costs less than a hit, a block at the attacker's
	// height pays back, and time pays back once the pressure stops.
	MaxStructure = 12
	HitStructure = 1
	// A run of hits costs more as it goes: each hit after the first in the same
	// run costs a step more than the one before, up to a ceiling. HitRunCapped
	// is the run length at which the cost has reached that ceiling - derived,
	// so retuning the step or the ceiling cannot leave the two disagreeing.
	HitStructureStep   = 0.5
	HitStructureMax    = 3
	HitRunCapped       = uint8((HitStructureMax - HitStructure) / HitStructureStep)
	BlockAllStructure  = 0.4
	BlockStructure     = 0.5
	StructureRegen     = 1  // per second, once calm
	StructureCalmTicks = 90 // 1.5s without a loss before it refills
	// What a kill pays the winner back. Two hits' worth: enough to be felt
	// going into whoever is next and nowhere near a heal, since a fight is
	// twelve hits long. Derived from the cost of a hit, so retuning that
	// carries the reward along with it.
	KillStructure = 2 * HitStructure

	// Energy is what a fighter spends to attack, counted in whole points.
	// Every swing costs the same, a block at the attacker's height pays back,
	// and time pays back once the fighter stops swinging - faster once nothing
	// is swinging at it either.
	MaxEnergy          = 300
	EnergySwing        = 12 // twenty-five swings from a full pool
	EnergyBlockRestore = 15 // more than a swing costs: one read pays for the answer
	// Refilling is slow enough to be felt: an empty pool takes this long to come
	// back once the fighter is calm. The tick gap is derived from it, so
	// retuning the pool cannot quietly retune the rate along with it.
	EnergyRefillSeconds = 10
	EnergyRegenTicks    = 60 * EnergyRefillSeconds / MaxEnergy // a point every 2 ticks, 30 per second
	EnergyCalmTicks     = 90                                   // 1.5s without a swing before it refills
	// A fighter nobody is trading with refills at twice the rate: the slow one
	// is what a fighter still under attack gets, and getting out of range long
	// enough for the attacks to stop is what buys the fast one.
	EnergyFreeRefillSeconds = 5
	EnergyFreeRegenTicks    = 60 * EnergyFreeRefillSeconds / MaxEnergy // a point every tick, 60 per second
	EnergyUntouchedTicks    = 120                                      // 2s without an attack reaching it
	// A running jump costs a third of the pool: three from full, and the fourth
	// has to be earned. Derived from MaxEnergy, so retuning the pool carries it
	// along.
	EnergyRunJump = MaxEnergy / 3

	StunEveryHits = 3
	// StunKnockback throws a stunned fighter about 20 px. It was 150 while
	// position was cut to whole pixels, which lost part of every tick of the
	// throw; with the fraction kept, 150 throws 33 px.
	StunKnockback   = 107
	FlinchHoldTicks = 12

	// A shove is what a standing fighter does about another on its back: the
	// one behind can swing at it and it cannot swing back. It turns on the
	// crowder and throws it off, costing it nothing but its ground.
	PushSeconds = 0.5 // three ticks a frame
	PushFrames  = 10
	PushFrame   = 3 // the frame both arms are out: where the shove lands
	// How far the shove moves the crowder. It starts 0 to 13px away, so this
	// must be at least 14 to clear the 14px overlap and at most 21 to stay
	// inside the 35px a swing reaches.
	PushDistance = 16

	// A clash is two fighters opening low swings at each other in the same
	// instant: instead of one guarding the other, the two trade blows in a
	// short exchange both play together. The window is in world ticks, and
	// well inside the ~13 ticks a low swing takes to reach its hit frame.
	ClashWindowTicks = 6 // 100 ms
	ClashStructure   = 0.5
	ClashSeconds     = 0.91 // 70 ms a frame; to be tuned by eye in play
	ClashLeftFrames  = 13
	ClashRightFrames = 13
	// ClashGap is how far apart the two stand while clashing, the distance the
	// two sheets' blows are drawn to meet at. A feel number, tuned in play.
	ClashGap = 24

	// Teams: fighters on the same one are allies - a bot does not pick an ally to
	// fight, and a swing passes through one. NoTeam is nobody's ally, not even
	// another fighter with no team, which is what a human is: every bot attacks
	// it, and humans can still fight each other. Every bot starts on BotTeam.
	NoTeam  = 0
	BotTeam = 1

	// One in three: the odds a bot answers a landed hit by moving its guard to
	// the height that hit it. A feel number, to be turned by eye in play.
	GuardAdaptOdds = 3

	DeathFlickerTicks = 36 // ~0.6s of flickering after the fall animation
	DeathFlickerSpan  = 4  // ticks per visible or hidden span
	RespawnTicks      = 120
)

type ActorState uint8

var attackFrames = map[string]uint8{
	"resources/threefour/attack-high.png":    5,
	"resources/threefour/attack-high-2.png":  3,
	"resources/threefour/attack-high-3.png":  3,
	"resources/threefour/attack-mid.png":     5,
	"resources/threefour/attack-mid-2.png":   3,
	"resources/threefour/attack-mid-3.png":   3,
	"resources/threefour/attack-low.png":     5,
	"resources/threefour/attack-low-2.png":   3,
	"resources/threefour/attack-low-3.png":   3,
	"resources/threefour/updown-atk-0-1.png": 4,
	"resources/threefour/updown-atk-0-2.png": 4,
	"resources/threefour/updown-atk-1-2.png": 4,
	"resources/threefour/updown-atk-1-0.png": 4,
	"resources/threefour/updown-atk-2-1.png": 4,
	"resources/threefour/updown-atk-2-0.png": 4,
}

const (
	StateIdle ActorState = iota
	StateRun
	StateAttack
	StateAttacked
	StateBlock
	StateJumpStart
	StateJump
	StateClimb
	StateTurnover
	StateDeath
	// appended, never inserted: the enum's values are what the snapshot
	// carries, so slotting a state in beside its relatives would renumber
	// every state below it for anyone drawing from an older build
	StateRunJump
	StateFall
	StatePush
	StateClash
	// stateCount is one past the last state. A replicated state is checked
	// against it, so appending a state cannot leave that check behind - it was
	// left at the death when the run-jump and the fall were appended, and every
	// snapshot carrying either was thrown away as malformed.
	stateCount
)

const (
	AttackFrame = 4
)

// presentation events, consumed (and cleared) by the renderer
const (
	EvAttackStart uint8 = 1 << iota
	EvHit
	EvBlocked
	EvNoEnergy
)

// Time manipulation records a fighter by copying its value, and restores it by
// copying that back. So an Actor holds no slice or map, and no pointer to
// anything that changes, other than the fighters and world it belongs to: a
// recorded frame would share it with the live match.
type Actor struct {
	ID               uint16
	W                *World
	Remote           bool // driven by snapshot interpolation, not simulated
	Pos              rl.Vector2
	PrevPos          rl.Vector2
	InterpolatedPos  rl.Vector2
	MoveToPos        rl.Vector2
	Speed            rl.Vector2
	Energy           int32
	Acc              float32
	Friction         float32
	Orientation      int8
	State            ActorState
	CollisionBox     rl.Rectangle
	AttackBoxHigh    rl.Rectangle
	AttackBoxMid     rl.Rectangle
	AttackBoxLow     rl.Rectangle
	OnGround         bool
	IsPlayer         bool
	Input            Input
	PrevUpdown       uint8
	Updown           uint8
	Ticks            uint32
	FreezeTicks      uint8
	BufAttack        bool // attack pressed while frozen; handed to the first live tick
	BufJump          bool // jump pressed while frozen; handed to the first live tick
	Events           uint8
	AnimStartTick    uint32
	AnimTop          *string
	Anim             string
	AnimDuration     float64
	AnimYOffset      int8
	AnimFrame        uint8
	AnimTotalFrames  uint8
	AnimCurrentFrame uint8
	AnimOnce         bool
	AnimReverse      bool
	AnimFinished     bool
	AnimWhite        bool
	Attacker         *Actor
	AttackQueued     bool
	AttackCount      uint8
	CanClimbRight    bool
	CanClimbLeft     bool
	Structure        float32
	HitCount         uint8   // landed hits since the last stun or recovery
	HitRun           uint8   // hits landed in a row; unlike HitCount a stun does not clear it
	Blocks           uint8   // blows this fighter has stopped, counted round; the view hears a change as a block
	CalmTicks        uint16  // ticks since structure was last lost
	EnergyCalm       uint16  // ticks since energy was last spent
	EnergyUntouched  uint16  // ticks since an attack last reached it, hit or blocked
	FlinchTicks      uint8   // ticks a hit that did not stun keeps the fighter flickering
	SwingHit         bool    // this fighter's current swing has already landed
	SwingTick        uint32  // the world tick this fighter last opened a swing; the authority's alone
	PushDone         bool    // this fighter's current shove has already moved someone
	PushStuck        *Actor  // the fighter its last shove landed on; not shoved again until the two are apart
	ShoveLeft        uint8   // pixels a shove still has to carry this fighter
	ShoveDir         int8    // which way that shove is carrying it
	Dying            bool    // dead and waiting to be taken out of the world
	DeathTicks       uint16  // ticks since the death animation finished
	DenyUntil        float64 // game time a refused swing stops flashing; local only, never replicated
	Owner            uint16  // network player this fighter belongs to; 0 for bots
	Team             uint8   // fighters on the same team do not fight; NoTeam is on nobody's
	// The two places at this fighter: the bot fighting it from the left, and
	// the one fighting it from the right. Two bots engage a fighter and no
	// more; the rest wait in line. Like the bot's mind below, this is worked
	// out by whichever world is simulating the match and is never replicated -
	// a client is told where the bots went, not whose turn it is. Time
	// manipulation records and restores it with the rest of the fighter.
	Engaged [2]*Actor
	// What the bot has in mind. Never replicated: it is worked out afresh by
	// whichever world is actually simulating the match, and a client is told
	// what the bot did, not what it was thinking. Time manipulation records and
	// restores it with the rest of the fighter. Zero is
	// a bot with nothing in mind, so a fighter spawning or respawning starts
	// with a clean one for free.
	Bot botMind
}

// Input is one simulation tick of player intent; it is also the network command.
type Input struct {
	Seq      uint32
	Left     bool
	Right    bool
	Attack   bool
	Jump     bool
	FaceLeft bool
	Updown   uint8
	ViewTick uint32
	ViewFrac uint8
}

func NewActor(pos rl.Vector2, isPlayer bool) *Actor {
	return &Actor{
		Pos:       pos,
		Speed:     rl.Vector2{},
		Acc:       DefaultAcc,
		Friction:  DefaultFriction,
		State:     StateIdle,
		Structure: MaxStructure,
		Energy:    MaxEnergy,
		// nothing has swung at it yet, so it refills at the free rate from the
		// first tick rather than having to wait the pressure out
		EnergyUntouched: EnergyUntouchedTicks,
		Orientation:     1,
		OnGround:        false,
		CollisionBox: rl.Rectangle{
			X:      14,
			Y:      6,
			Width:  14,
			Height: 42,
		},
		AttackBoxHigh: rl.Rectangle{
			0, 5, 24, 6,
		},
		AttackBoxMid: rl.Rectangle{
			0, 14, 24, 6,
		},
		AttackBoxLow: rl.Rectangle{
			0, 24, 24, 6,
		},
		IsPlayer:        isPlayer,
		Team:            teamFor(isPlayer),
		Anim:            "resources/threefour/updown.png",
		AnimTotalFrames: 1,
		AnimDuration:    1,
		AnimFinished:    false,
		PrevUpdown:      2,
		Updown:          0,
	}
}

// teamFor is the team a fighter starts on: every bot on BotTeam, every human on
// none.
func teamFor(isPlayer bool) uint8 {
	if isPlayer {
		return NoTeam
	}
	return BotTeam
}

// allyOf reports whether a and b are on the same team.
func (a *Actor) allyOf(b *Actor) bool {
	return a.Team != NoTeam && a.Team == b.Team
}

// placeIndex is which of a fighter's two places belongs to a side: 0 for the
// left, 1 for the right.
func placeIndex(side int8) int {
	if side < 0 {
		return 0
	}
	return 1
}

// engageSide is which side of t this fighter is on. Two fighters on the same
// pixel column have no side to be on, and there the facing decides - the same
// answer pushDir gives a shove thrown at that distance.
func (a *Actor) engageSide(t *Actor) int8 {
	if d := a.Pos.X - t.Pos.X; d != 0 {
		return facing(d)
	}
	return a.Orientation
}

// holdsPlace reports whether this bot is standing in the place it last took.
// The fighter's own pair is the only record of who holds what; the bot's side
// only says which of the two to look in, so the two cannot drift apart.
func (a *Actor) holdsPlace() bool {
	t := a.Bot.target
	return t != nil && t.Engaged[placeIndex(a.Bot.side)] == a
}

// releasePlace gives up the place this bot holds, if it holds one.
//
// It does not forget the target. A bot whose turn at the front is over is still
// fighting the same fighter - it goes to the back of the line and waits for
// another - and leaveFight is what ends that.
func (a *Actor) releasePlace() {
	t := a.Bot.target
	if t == nil {
		return
	}
	if i := placeIndex(a.Bot.side); t.Engaged[i] == a {
		t.Engaged[i] = nil
	}
	a.Bot.holdTicks = 0
}

// leaveFight gives up both the place and the fighter it was held at: this bot
// is no longer in that fight at all, at the front of it or in the line.
func (a *Actor) leaveFight() {
	a.releasePlace()
	a.Bot.target, a.Bot.side = nil, 0
}

// Update advances the actor by one fixed tick using a.Input.
func (a *Actor) Update(dt float32) {
	a.PrevPos = a.Pos
	// Hit-stop holds input, it does not eat it. A frozen tick puts the press
	// aside; the first live tick takes it back and clears it. So an attack or
	// jump read below may be up to HitStopTicks old - none of the state arms
	// that read Input can tell, and none of them need to.
	if a.FreezeTicks > 0 {
		a.FreezeTicks--
		a.BufAttack = a.BufAttack || a.Input.Attack
		a.BufJump = a.BufJump || a.Input.Jump
		return
	}
	if a.BufAttack || a.BufJump {
		// folded in for this tick only, then put back: Input is the command
		// that arrived, and a held press that this tick does not act on has
		// been offered and is spent. Leaving it in would let a stun, which
		// reads no input, hand the press on to whatever state follows it.
		cmd := a.Input
		defer func() { a.Input = cmd }()
		a.Input.Attack = a.Input.Attack || a.BufAttack
		a.Input.Jump = a.Input.Jump || a.BufJump
		a.BufAttack, a.BufJump = false, false
	}

	prevOrientation := a.Orientation
	a.Ticks++
	defer a.advanceAnim()

	if a.State == StateDeath {
		a.updateDeath()
		return
	}
	a.regenStructure()
	a.regenEnergy()
	if a.FlinchTicks > 0 {
		a.FlinchTicks--
		if a.FlinchTicks == 0 {
			a.AnimWhite = false
		}
	}

	a.checkCollision()
	a.updateHorSpeed(dt)
	a.updateVertSpeed(dt)
	a.updateFalling()

	// Two fighters who are not allies cannot stand in one another: whichever is
	// on its feet and free - standing, walking or running - turns on the other
	// and shoves it off. One already swinging, blocking, stunned or still
	// flinching from a hit finishes that first, and one guarding everything
	// stands its ground. The first to qualify shoves, and the stun it throws
	// keeps the other from shoving back.
	if (a.State == StateIdle || a.State == StateRun) && a.OnGround && a.Updown != 3 && a.FlinchTicks == 0 {
		if other := a.crowding(); other != nil {
			a.enterPushState(other)
		}
	}

	switch a.State {
	case StateClimb:
		a.Speed = rl.Vector2{}
		if a.AnimFinished {
			a.enterIdleState()
		}
	case StateIdle:
		prevUpdown := a.Updown
		a.updateUpdown()
		if prevUpdown == 3 && a.Updown != 3 {
			a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
		} else if a.Updown == 3 {
			a.SetAnimFrame("resources/threefour/block.png", 0)
		} else if absSpeed := abs(a.Speed.X); absSpeed > RunReach {
			a.enterRunState()
			return
		} else if absSpeed > 0.1 {
			reverse := true
			if a.Speed.X > 0 && !a.MouseOnLeft() || a.Speed.X < 0 && a.MouseOnLeft() {
				reverse = false
			}
			reset := a.AnimReverse != reverse
			a.SetAnim("resources/threefour/walk.png", AnimOptions{reset: reset, animDuration: 0.6, totalFrames: 9, reverse: reverse, startFrame: 9})
		} else if prevUpdown != a.Updown {
			anim := fmt.Sprintf("resources/threefour/updown-%d-%d.png", prevUpdown, a.Updown)
			a.SetAnim(anim, AnimOptions{animDuration: 0.1, totalFrames: 4, once: true})
		} else if a.AnimFinished && strings.Contains(a.Anim, "updown-") {
			a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
		} else if absSpeed < 0.1 && a.Anim == "resources/threefour/walk.png" {
			a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
		}
		if a.Input.Attack && a.Updown != 3 && a.enterAttackState() {
			return
		}
		if a.Input.Jump {
			a.enterJumpStartState()
			return
		}
	case StateRun:
		if prevOrientation != a.Orientation {
			a.enterTurnoverState()
		}
		absSpeed := abs(a.Speed.X)
		if absSpeed < 0.001 {
			a.enterIdleState()
			return
		}
		// no return: the take-off tick must still apply its movement, or the
		// fighter is left standing on the floor and the next collision check
		// cancels the jump. The standing jump leaves the ground the same way.
		if a.Input.Jump && a.enterRunJumpState() {
			// nothing else this tick; a refused jump falls through and the
			// fighter goes on running
		} else if a.Input.Attack {
			a.enterAttackState()
		}
	case StateAttack:
		a.Speed.X = clamp(a.Speed.X, -AttackMoveSpeed, AttackMoveSpeed)
		a.updateAttack()
	case StatePush:
		a.updatePush()
	case StateClash:
		// the exchange is played out, not steered: nothing asked for during it
		// is carried out, then or after
		a.Speed.X = 0
		if a.AnimFinished {
			a.endClash()
		}
	case StateAttacked:
		// no AttackDash cap here: that is the lunge of the fighter doing the
		// swinging, and capping the one being hit threw StunKnockback away.
		// The lockout says what a stunned fighter may do, not how far the hit
		// carries it; friction ends the throw, as it does every other.
		a.updateAttacked()
	case StateBlock:
		a.updateBlock()
	case StateJumpStart:
		if a.AnimFinished {
			a.enterJumpState(dt)
		}
	case StateJump:
		if (a.CanClimbRight && a.Orientation == 1) || (a.CanClimbLeft && a.Orientation == -1) {
			a.exitJumpState()
			a.enterClimbState()
		} else if a.OnGround {
			a.exitJumpState()
			a.enterIdleState()
		}
	case StateRunJump:
		// still on the ground until the push-off frame arrives; the fighter is
		// committed from the tick it pressed jump, so the button is not read
		// again here
		if a.OnGround {
			if a.AnimCurrentFrame >= RunJumpLaunchFrame {
				a.launchRunJump(dt)
			}
			break
		}
		if (a.CanClimbRight && a.Orientation == 1) || (a.CanClimbLeft && a.Orientation == -1) {
			a.exitJumpState()
			a.enterClimbState()
		}
	case StateFall:
		// a fall catches a ledge on the same terms a jump does: falling costs
		// the fighter nothing it had on the way up
		if (a.CanClimbRight && a.Orientation == 1) || (a.CanClimbLeft && a.Orientation == -1) {
			a.exitJumpState()
			a.enterClimbState()
		} else if a.OnGround {
			a.exitJumpState()
			a.land()
		}
	case StateTurnover:
		if a.AnimFinished {
			if abs(a.Speed.X) < 0.001 {
				a.enterIdleState()
			} else if a.Input.Left || a.Input.Right {
				a.Speed.X = RunReach * float32(a.Orientation)
				a.enterRunState()
			} else {
				a.enterIdleState()
			}
		}
	}

	// X keeps its fraction from tick to tick, so any speed moves the fighter;
	// it is rounded to a whole pixel only where it is drawn. Cutting it to a
	// whole pixel here stopped everything under 60 px/s, and cutting toward
	// zero made a walk faster one way than the other.
	fromX := a.Pos.X
	a.Pos.X += a.Speed.X * dt
	a.blockSides(fromX)
	if a.ShoveLeft > 0 {
		a.glide()
	}
	a.Pos.Y += a.Speed.Y * dt
	a.Pos.Y = float32(int(a.Pos.Y))
}

// advanceAnim derives the current animation frame from the actor's own tick counter.
func (a *Actor) advanceAnim() {
	elapsed := float64(a.Ticks-a.AnimStartTick) * FixedTimeStep
	a.AnimCurrentFrame, a.AnimFinished = animFrame(a.AnimOnce, a.AnimReverse, a.AnimFrame, a.AnimTotalFrames, a.AnimDuration, elapsed)
}

func animFrame(once, reverse bool, start, total uint8, duration, elapsed float64) (uint8, bool) {
	if total == 0 || duration <= 0 {
		return start, once
	}
	frameTime := duration / float64(total)
	if once {
		frame := int(elapsed / frameTime)
		finished := frame >= int(total)
		return uint8(clamp(frame, 0, int(total)-1)), finished
	}
	elapsed = math.Mod(max(elapsed, 0), duration)
	frame := uint8(clamp(int(elapsed/frameTime), 0, int(total)-1))
	if reverse {
		return start - frame, false
	}
	return start + frame, false
}

func (a *Actor) UpdatePhysics() {}

// drawPos is the sprite's position rounded to whole world pixels. Interpolation
// and the client's error correction leave sub-pixel positions, which the
// integer camera zoom turns into texel rows of uneven width - a sprite drawn at
// x.5 shimmers as it moves.
func (a *Actor) drawPos() rl.Vector2 {
	return rl.Vector2{X: roundf(a.InterpolatedPos.X), Y: roundf(a.InterpolatedPos.Y)}
}

func (a *Actor) Draw() {
	if a.deathHidden() {
		return
	}
	if a.AnimOnce {
		a.DrawAnimOnce()
	} else {
		a.DrawAnim()
	}

	if DevTools && Debug {
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      a.Pos.X,
			Y:      a.Pos.Y,
			Width:  48,
			Height: 48,
		}, 1, rl.Blue)
		rl.DrawRectangleLinesEx(a.collisionRect(), 1, rl.SkyBlue)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      a.Pos.X + a.AttackBoxHigh.X + 12*float32(a.Orientation+1),
			Y:      a.Pos.Y + a.AttackBoxHigh.Y,
			Width:  a.AttackBoxHigh.Width,
			Height: a.AttackBoxHigh.Height,
		}, 1, rl.Red)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      a.Pos.X + a.AttackBoxMid.X + 12*float32(a.Orientation+1),
			Y:      a.Pos.Y + a.AttackBoxMid.Y,
			Width:  a.AttackBoxMid.Width,
			Height: a.AttackBoxMid.Height,
		}, 1, rl.Red)
		rl.DrawRectangleLinesEx(rl.Rectangle{
			X:      a.Pos.X + a.AttackBoxLow.X + 12*float32(a.Orientation+1),
			Y:      a.Pos.Y + a.AttackBoxLow.Y,
			Width:  a.AttackBoxLow.Width,
			Height: a.AttackBoxLow.Height,
		}, 1, rl.Red)
		if a.IsPlayer {
			rl.DrawText(fmt.Sprintf("Time: %s", timeModeName()), int32(a.Pos.X), int32(a.Pos.Y), 2, rl.White)
			rl.DrawText(fmt.Sprintf("Frame: %d of %d", rewindCursor-rewindOldest()+1, rewindCount), int32(a.Pos.X), int32(a.Pos.Y)-8, 2, rl.White)
			rl.DrawText(fmt.Sprintf("X, Y: %d %d", int32(a.Pos.X), int32(a.Pos.Y)), int32(a.Pos.X), int32(a.Pos.Y)-16, 2, rl.White)
			rl.DrawText(fmt.Sprintf("PrevUpdown, Updown: %d %d", a.PrevUpdown, a.Updown), int32(a.Pos.X), int32(a.Pos.Y)-24, 2, rl.White)
			rl.DrawText(fmt.Sprintf("Frame, Finished: %d %t", a.AnimCurrentFrame, a.AnimFinished), int32(a.Pos.X), int32(a.Pos.Y)-32, 2, rl.White)
		}
		px := a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2
		rl.DrawPixel(int32(px), int32(a.Pos.Y+a.CollisionBox.Y+a.CollisionBox.Height), rl.Red)
		rl.DrawText(fmt.Sprintf("State: %d", a.State), int32(a.Pos.X), int32(a.Pos.Y+a.CollisionBox.Y+a.CollisionBox.Height), 2, rl.White)
		rl.DrawText(fmt.Sprintf("Struct: %.1f hits: %d", a.Structure, a.HitCount), int32(a.Pos.X), int32(a.Pos.Y+a.CollisionBox.Y+a.CollisionBox.Height)+8, 2, rl.White)
	}

}

func (a *Actor) collisionRect() rl.Rectangle {
	return rl.Rectangle{
		X:      a.Pos.X + a.CollisionBox.X,
		Y:      a.Pos.Y + a.CollisionBox.Y,
		Width:  a.CollisionBox.Width,
		Height: a.CollisionBox.Height,
	}
}

func (a *Actor) SetAnimFrame(anim string, offset uint8) {
	a.AnimStartTick = a.Ticks
	a.Anim = anim
	a.AnimDuration = 1
	a.AnimTotalFrames = 1
	a.AnimFinished = false
	a.AnimOnce = false
	a.AnimFrame = offset
}

type AnimOptions struct {
	animDuration float64
	totalFrames  uint8
	animYOffset  int8
	startFrame   uint8
	once         bool
	reset        bool
	reverse      bool
}

func (a *Actor) SetAnim(anim string, o AnimOptions) {
	if a.Anim == anim && !o.reset {
		return
	}
	a.AnimStartTick = a.Ticks
	a.Anim = anim
	a.AnimFinished = false
	a.AnimDuration = o.animDuration
	a.AnimTotalFrames = o.totalFrames
	a.AnimOnce = o.once
	a.AnimReverse = o.reverse
	a.AnimFrame = o.startFrame
	a.AnimYOffset = o.animYOffset
}

func (a *Actor) DrawFrame() {
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}

	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(a.AnimCurrentFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: float32(tex.Height),
	}, rl.Vector2{
		X: a.drawPos().X,
		Y: a.drawPos().Y + float32(a.AnimYOffset),
	}, rl.White)
}

// spriteDraw is the texture, source rectangle and position this fighter's
// sprite is drawn with on this frame. The outline and rim passes draw the same
// sprite a second time, so they must read it from one place or they would drift
// and trace a pose the fighter is not in.
func (a *Actor) spriteDraw() (rl.Texture2D, rl.Rectangle, rl.Vector2) {
	tex := resources[a.Anim]
	pos := a.drawPos()
	height := float32(48)
	if a.AnimOnce {
		height = float32(tex.Height)
		pos.Y += float32(a.AnimYOffset)
	}
	return tex, rl.Rectangle{
		X:      float32(a.AnimCurrentFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: height,
	}, pos
}

// topAnim is the animation drawn over this fighter's body on this frame, if it
// has one. The walk is drawn in two layers: the body walks and the arms are a
// second sheet chosen by the stance, so the fighter holds its guard while it
// moves. Block-all has no arms of its own and is drawn as the body alone.
//
// TODO the walk's rule lives here rather than in the state machine; when
// AnimTop is set where the animation is, this falls away and AnimTop stays.
func (a *Actor) topAnim() (string, bool) {
	if a.AnimTop != nil {
		return *a.AnimTop, true
	}
	if !strings.HasSuffix(a.Anim, "walk.png") {
		return "", false
	}
	switch a.Updown {
	case 0:
		return "resources/threefour/walk-bot.png", true
	case 1:
		return "resources/threefour/walk-mid.png", true
	case 2:
		return "resources/threefour/walk-top.png", true
	}
	return "", false
}

// topDraw is the second sprite drawn over this fighter's body on this frame: the
// same frame of the top layer's sheet, at the same place as the body. The
// drawing and the rim both take it from here, so the rim cannot trace a
// silhouette the fighter is not showing. ok is false when the fighter has no
// top layer.
func (a *Actor) topDraw() (tex rl.Texture2D, src rl.Rectangle, pos rl.Vector2, ok bool) {
	anim, has := a.topAnim()
	if !has {
		return tex, src, pos, false
	}
	tex, ok = resources[anim]
	if !ok {
		return tex, src, pos, false
	}
	_, src, pos = a.spriteDraw()
	return tex, src, pos, true
}

func (a *Actor) DrawAnimOnce() {
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}

	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(a.AnimCurrentFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: float32(tex.Height),
	}, rl.Vector2{
		X: a.drawPos().X,
		Y: a.drawPos().Y + float32(a.AnimYOffset),
	}, rl.White)
}

func (a *Actor) DrawAnim() {
	tex, ok := resources[a.Anim]
	if !ok {
		panic(a.Anim)
	}

	rl.DrawTextureRec(tex, rl.Rectangle{
		X:      float32(a.AnimCurrentFrame) * 48,
		Y:      0,
		Width:  48 * float32(a.Orientation),
		Height: 48,
	}, a.drawPos(), rl.White)

	if top, src, pos, ok := a.topDraw(); ok {
		rl.DrawTextureRec(top, src, pos, rl.White)
	}
}

func (a *Actor) updateHorSpeed(dt float32) {
	var movePressed bool
	prevOrientation := a.Orientation
	canMove := a.State != StateBlock && a.State != StateAttacked && a.State != StatePush && a.State != StateClash && a.Updown != 3
	// a swinging fighter moves, slowly (the attack case caps it), but keeps
	// facing the way it swings
	turn := a.State != StateAttack
	if a.Input.Right && !(a.State == StateTurnover && prevOrientation == 1) && canMove {
		movePressed = true
		if abs(a.Speed.X) < 0.001 {
			a.Speed.X -= 12 * a.Acc * dt
		} else {
			a.Speed.X -= a.Acc * dt
		}
		if turn {
			a.face(-1)
		}
	}
	if a.Input.Left && !(a.State == StateTurnover && prevOrientation == -1) && canMove {
		movePressed = true
		if abs(a.Speed.X) < 0.001 {
			a.Speed.X += 12 * a.Acc * dt
		} else {
			a.Speed.X += a.Acc * dt
		}
		if turn {
			a.face(1)
		}
	}
	if a.State == StateIdle || a.State == StateAttack && a.AnimCurrentFrame == 0 {
		if a.IsPlayer {
			if a.MouseOnLeft() {
				a.face(-1)
			} else {
				a.face(1)
			}
		}
	}

	if prevOrientation != a.Orientation {
		a.Speed.X /= 3
	}

	if !movePressed {
		a.Speed.X = moveTo0(a.Speed.X, a.Friction*dt) // friction
		// a walk let go of stops at the pace the old whole-pixel floor stopped
		// it at; a speed something else gave the fighter - a flinch, a shove -
		// runs out under friction instead
		own := (a.State == StateIdle || a.State == StateRun) && a.FlinchTicks == 0 && a.ShoveLeft == 0
		if own && abs(a.Speed.X) < WalkStopSpeed {
			a.Speed.X = 0
		}
	}
	a.Speed.X = clamp(a.Speed.X, -MaxSpeed, MaxSpeed)
}

func (a *Actor) updateVertSpeed(dt float32) {
	if !a.OnGround {
		a.Speed.Y += Gravity * dt
	}
	a.Speed.Y = clamp(a.Speed.Y, MaxJumpSpeed, MaxFallSpeed)
}

// headBlocked reports whether a collision rectangle contains the actor's head.
// The climb checks in checkCollision need this up front: the solids follow
// objectsIndexed, sorted by descending X, so a neighbouring tile of a ceiling row is visited
// before the tile the head is actually inside of.
func (a *Actor) headBlocked(pColCenterX float32) bool {
	pColUpY := a.Pos.Y + a.CollisionBox.Y
	for _, c := range LE.solids {
		oColY := c.Y
		oColX := c.X
		if oColX <= pColCenterX && oColX+c.Width >= pColCenterX &&
			pColUpY >= oColY && pColUpY <= oColY+c.Height {
			return true
		}
	}
	return false
}

func (a *Actor) checkCollision() {
	a.OnGround = false
	pColDownY := a.Pos.Y + a.CollisionBox.Y + a.CollisionBox.Height
	pColCenterX := a.Pos.X + a.CollisionBox.X + a.CollisionBox.Width/2
	pPosX := a.Pos.X
	hitTop := a.headBlocked(pColCenterX)
	for _, c := range LE.solids {
		oColY := c.Y
		oColHeight := c.Height
		oColX := c.X
		oColWidth := c.Width

		// The left edge claims the fighter and the right edge does not, so a
		// tile boundary belongs to exactly one tile. With both edges exclusive
		// a fighter whose centre landed exactly on a seam was claimed by
		// neither, left the ground for that one tick and was put into the fall
		// - which, walking an unbroken floor, cancelled any jump winding up.
		if pColCenterX >= oColX && pColCenterX < oColX+oColWidth &&
			pColDownY+1 > oColY && pColDownY < oColY+oColHeight {
			a.OnGround = true
			a.Speed.Y = 0
			a.Pos.Y = float32(int32(math.Ceil(float64(oColY)) - 48))
			continue
		}
		pColUpY := a.Pos.Y + a.CollisionBox.Y
		// a ledge is caught going either way: rising into it on a jump, or
		// dropping past it on a fall
		airborne := a.Speed.Y != 0
		const climbOffset = 32
		if oColX <= pColCenterX && oColX+oColWidth >= pColCenterX &&
			pColUpY >= oColY && pColUpY <= oColY+oColHeight {
			hitTop = true
			a.CanClimbRight = false
			a.CanClimbLeft = false
			a.Speed.Y = 0
			a.Pos.X = pPosX
			a.Pos.Y = float32(int32(oColY + oColHeight))
			continue
		} else if pColCenterX+climbOffset >= oColX && oColX >= pColCenterX &&
			pColUpY >= oColY && pColUpY <= oColY+oColHeight && a.Orientation == 1 && airborne && !hitTop {
			a.CanClimbRight = true
			a.Pos.X = float32(int32(oColX) - 20)
			a.Pos.Y = float32(int32(math.Ceil(float64(oColY)) - 48 + 1))
			continue
		} else if pColCenterX >= oColX+oColWidth && oColX+oColWidth >= pColCenterX-climbOffset &&
			pColUpY >= oColY && pColUpY <= oColY+oColHeight && a.Orientation == -1 && airborne && !hitTop {
			a.CanClimbLeft = true
			a.Pos.X = oColX + oColWidth - 24
			a.Pos.Y = float32(int32(math.Ceil(float64(oColY)) - 48 + 1))
			continue
		}
	}
}

// isWall reports whether solid c stands in the fighter's way sideways: some of
// it is level with the body, from the head down to a step's height above the
// feet. Anything lower is a step, walked onto rather than stopped at.
func (a *Actor) isWall(c rl.Rectangle) bool {
	head := a.Pos.Y + a.CollisionBox.Y
	feet := head + a.CollisionBox.Height
	return c.Y < feet-StepHeight && c.Y+c.Height > head
}

// blockSides cuts a sideways move short, from fromX to where the fighter now
// is, at the first wall face its collision box crossed, and takes away any
// speed carrying it into that wall. It stops the crossing rather than the
// overlap, so no wall is too thin to stop a fighter at any speed it reaches,
// and a fighter already inside a wall is never pushed out of it - only kept
// from going deeper.
// stopRun takes a fighter out of the run when a wall has stopped it: a run with
// nowhere to go is not a run, and the key held against the wall walks in place
// instead of running in place. It is the same end the run comes to when it
// slows to a stop of its own accord.
//
// Only a fighter on the ground. One steering into a wall in the air is stopped
// beside it and goes on rising or falling, which is a jump or a fall's business,
// not the run's.
func (a *Actor) stopRun() {
	if a.State == StateRun && a.OnGround {
		a.enterIdleState()
	}
}

func (a *Actor) blockSides(fromX float32) {
	l, w := a.CollisionBox.X, a.CollisionBox.Width
	fromL, toL := fromX+l, a.Pos.X+l
	switch {
	case toL > fromL:
		right := toL + w
		for _, c := range LE.solids {
			if c.X > fromL && a.isWall(c) {
				right = min(right, max(c.X, fromL+w))
			}
		}
		if right < toL+w {
			a.Pos.X = right - w - l
			a.Speed.X = min(a.Speed.X, 0)
			a.stopRun()
		}
	case toL < fromL:
		left := toL
		for _, c := range LE.solids {
			if c.X+c.Width < fromL+w && a.isWall(c) {
				left = max(left, min(c.X+c.Width, fromL))
			}
		}
		if left > toL {
			a.Pos.X = left - l
			a.Speed.X = max(a.Speed.X, 0)
			a.stopRun()
		}
	}
}

// edge-triggered buttons are latched per rendered frame and consumed by the next tick
var pressedAttack, pressedJump bool

func latchInput() {
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		pressedAttack = true
	}
	if rl.IsKeyPressed(rl.KeyW) {
		pressedJump = true
	}
}

// matchHeld reports whether a paused mode is holding the match: edit mode, or
// time manipulation. No fixed step runs while it is true, so nothing would
// consume input latched for one. Every pause path answers this one question,
// rather than each keeping its own list of conditions.
func matchHeld() bool {
	return Editing || TimeMode != TimeOff
}

// clearLatchedInput drops input latched for a fixed step that is not going to
// run. Both paused modes call it as they end, so a click made while the match
// was held cannot fire the moment it resumes.
func clearLatchedInput() {
	pressedAttack, pressedJump = false, false
}

// sampleInput reads the local player's devices into a command for actor a.
func sampleInput(a *Actor) Input {
	in := Input{
		Right:  rl.IsKeyDown(rl.KeyA),
		Left:   rl.IsKeyDown(rl.KeyD),
		Attack: pressedAttack,
		Jump:   pressedJump,
	}
	pressedAttack, pressedJump = false, false

	curPos := rl.GetScreenToWorld2D(Cur.AtkPos, Cam.View())
	switch {
	case float32(curPos.Y) <= a.Pos.Y+a.CollisionBox.Y:
		in.Updown = 3
	case float32(curPos.Y) <= a.Pos.Y+16:
		in.Updown = 2
	case float32(curPos.Y) <= a.Pos.Y+24:
		in.Updown = 1
	default:
		in.Updown = 0
	}
	in.FaceLeft = float32(MouseWorldPos.X) < a.Pos.X+24
	return in
}

func (a *Actor) updateUpdown() {
	prevUpdown := a.Updown
	a.Updown = a.Input.Updown
	if a.Updown != prevUpdown {
		a.PrevUpdown = prevUpdown
	}
}

// enterAttackState starts a swing, if the fighter can pay for one. A swing that
// cannot be paid for does not happen: nothing is spent, nothing is entered, and
// the refusal is raised for the renderer to show.
func (a *Actor) enterAttackState() bool {
	if !a.canSwing() {
		a.Events |= EvNoEnergy
		return false
	}
	a.spendEnergy(EnergySwing)
	a.State = StateAttack
	if a.W != nil {
		a.SwingTick = a.W.Tick
	}
	a.stopFlicker()
	a.SwingHit = false
	a.Events |= EvAttackStart
	fromX := a.Pos.X
	a.Pos.X += float32(a.Orientation) * AttackDash
	a.blockSides(fromX)
	switch a.Updown {
	case 0:
		a.SetAnim("resources/threefour/attack-low.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	case 1:
		a.SetAnim("resources/threefour/attack-mid.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	case 2, 3:
		a.SetAnim("resources/threefour/attack-high.png", AnimOptions{animDuration: 0.3, totalFrames: 8, once: true})
	}
	return true
}

// spendStructure takes cost off the pool and reports whether that emptied it.
func (a *Actor) spendStructure(cost float32) bool {
	a.CalmTicks = 0
	a.Structure -= cost
	if a.Structure <= 0 {
		a.Structure = 0
		return true
	}
	return false
}

// restoreStructure puts structure back, never past the cap.
func (a *Actor) restoreStructure(amount float32) {
	a.Structure = min(a.Structure+amount, MaxStructure)
}

// face turns the fighter, and is the only way anything but a replayed or
// replicated record does: a climb or a death refuses the turn. The climb
// animation is what carries the fighter over the ledge it grabbed, and one that
// turned halfway up would mantle one way while facing the other. A broken
// fighter has stopped being part of the fight, so nothing it can still see
// turns it: the fall would spin on the spot while it played out. A clash is
// played out facing the other fighter, whatever either one asks. Every turn a
// player or a bot asks for goes through here so the rule cannot be missed by a
// new caller.
func (a *Actor) face(orientation int8) {
	if a.State == StateClimb || a.State == StateDeath || a.State == StateClash {
		return
	}
	a.Orientation = orientation
}

// canSwing reports whether the fighter can pay for a swing.
func (a *Actor) canSwing() bool { return a.Energy >= EnergySwing }

// canRunJump reports whether the fighter can pay for a running jump.
func (a *Actor) canRunJump() bool { return a.Energy >= EnergyRunJump }

// spendEnergy pays cost out of the pool. The calm wait starts again: spending
// is the only thing that stops the refill outright, whether it went on a swing
// or on leaving the ground at a run. Being attacked only slows it, which is
// regenEnergy's business.
func (a *Actor) spendEnergy(cost int32) {
	a.EnergyCalm = 0
	a.Energy = max(a.Energy-cost, 0)
}

// restoreEnergy puts energy back, never past the cap. It leaves the calm wait
// alone, so a block pays on top of whatever refilling was already under way.
func (a *Actor) restoreEnergy(amount int32) {
	a.Energy = min(a.Energy+amount, MaxEnergy)
}

// pressured marks the fighter as being attacked: an attack reached it, landed
// or guarded. Energy refills at the slow rate until attacks have stopped
// reaching it for the whole untouched wait.
func (a *Actor) pressured() { a.EnergyUntouched = 0 }

// regenStructure refills the pool once the fighter has been left alone long
// enough. The hit count goes with it: a run of hits is over once its target has
// had time to recover from it.
func (a *Actor) regenStructure() {
	if a.W != nil && !a.W.Authority {
		return // a predicted world displays structure, it does not grant it
	}
	if a.CalmTicks < StructureCalmTicks {
		a.CalmTicks++
		if a.CalmTicks == StructureCalmTicks {
			// both counts go: a fighter left alone this long is out from under
			// whatever was happening to it, the stun and the run alike
			a.HitCount = 0
			a.endHitRun()
		}
		return
	}
	a.restoreStructure(StructureRegen * float32(FixedTimeStep))
}

// regenEnergy refills the pool once the fighter has stopped swinging, at one of
// two rates. Swinging is the only thing that stops the refill outright; being
// attacked holds it to the slow rate, and it doubles once nothing has reached
// the fighter - hit or blocked - for the untouched wait. A fighter that is
// neither swinging nor being swung at is the one that gets its pool back fast.
//
// Unlike regenStructure this runs in a predicted world too. A client that spent
// but did not refill would refuse swings the server allows, and the server's
// swing would then arrive as an attack the player never saw start; all three
// inputs it is decided from, the tick counter and the two waits, are replicated
// so a client refills on the same tick the server does.
func (a *Actor) regenEnergy() {
	if a.EnergyUntouched < EnergyUntouchedTicks {
		a.EnergyUntouched++
	}
	if a.EnergyCalm < EnergyCalmTicks {
		a.EnergyCalm++
		return
	}
	gap := uint32(EnergyRegenTicks)
	if a.EnergyUntouched >= EnergyUntouchedTicks {
		gap = EnergyFreeRegenTicks
	}
	if a.Energy < MaxEnergy && a.Ticks%gap == 0 {
		a.restoreEnergy(1)
	}
}

// hitCost is what the next hit on this fighter costs: the base, plus a step for
// each hit already landed in the same run, never past the ceiling. It reads the
// run as it stands before the hit, so the first hit of a run costs the base.
func (a *Actor) hitCost() float32 {
	return min(HitStructure+HitStructureStep*float32(a.HitRun), HitStructureMax)
}

// endHitRun ends the run of hits against this fighter: the next one to land
// costs the base again. A block does this, and so does recovery.
func (a *Actor) endHitRun() { a.HitRun = 0 }

// takeHit resolves an attack that landed. Every hit costs structure and
// flinches, every third one stuns instead, and the one that empties the pool
// kills. It is one of the two places structure is spent.
func (a *Actor) takeHit(attacker *Actor) {
	attacker.SwingHit = true
	a.alerted(attacker)
	a.pressured()
	cost := a.hitCost()
	a.HitRun = min(a.HitRun+1, HitRunCapped)
	if a.spendStructure(cost) {
		a.enterDeathState(attacker)
		return
	}
	// A bot answers the hit a third of the time by guarding where it landed, so
	// one attack repeated is not reliable forever. Bots only, and explicitly:
	// a player's guard follows their aim, and writing it here - even for the
	// one tick before the cursor takes it back - would be read by any other
	// attacker whose hit is resolved later in the same actor loop.
	if !a.IsPlayer && rand.Int()%GuardAdaptOdds == 0 {
		if a.Updown != attacker.Updown {
			a.PrevUpdown = a.Updown
		}
		a.Updown = attacker.Updown
	}
	a.HitCount++
	if a.HitCount >= StunEveryHits {
		a.HitCount = 0
		a.enterAttackedState(attacker)
		return
	}
	a.enterFlinchState(attacker)
}

// confirmBlock resolves an attack that was guarded. A block at the attacker's
// height feeds the pool; block-all stops anything but pays for it, less than a
// hit would have cost.
func (a *Actor) confirmBlock(attacker *Actor) {
	attacker.SwingHit = true
	a.alerted(attacker)
	a.pressured()
	a.endHitRun() // any attack stopped is the run broken, block-all included
	a.Blocks++
	if a.Updown == 3 {
		if a.spendStructure(BlockAllStructure) {
			a.enterDeathState(attacker)
		}
		return
	}
	a.restoreStructure(BlockStructure)
	a.restoreEnergy(EnergyBlockRestore)
}

// hitFeedback is what every landed hit does, stunning or not. The hit
// animation is not part of it: only the stun is worth interrupting a fighter
// for, so a flinch is told apart by the flicker alone.
func (a *Actor) hitFeedback(attacker *Actor, knockback float32) {
	a.Attacker = attacker
	a.AnimWhite = true
	a.Events |= EvHit
	a.FreezeTicks = HitStopTicks
	attacker.FreezeTicks = HitStopTicks
	a.Speed.X = float32(attacker.Orientation) * knockback
}

// enterFlinchState is a hit that did not stun: the fighter keeps control and
// whatever it was doing, and only flickers for a moment.
//
// A shove is the exception: a hit before it lands is the answer to it, and the
// fighter is left standing with the shove not thrown.
func (a *Actor) enterFlinchState(attacker *Actor) {
	if a.State == StatePush {
		a.enterIdleState()
	}
	a.hitFeedback(attacker, PunchSpeed)
	a.FlinchTicks = FlinchHoldTicks
}

// enterAttackedState is the stun every third hit lands: a lockout until the
// animation finishes, and a longer throw back than a flinch. It is the only
// hit that plays the hit animation.
func (a *Actor) enterAttackedState(attacker *Actor) {
	a.State = StateAttacked
	a.stopFlicker()
	a.hitFeedback(attacker, StunKnockback)
	a.SetAnim("resources/threefour/hit-top.png", AnimOptions{animDuration: 0.3, totalFrames: 4, once: true})
}

// stopFlicker ends a flinch early: anything the fighter does of its own accord
// takes the flicker with it, or it would be left white for good.
func (a *Actor) stopFlicker() {
	a.FlinchTicks = 0
	a.AnimWhite = false
}

// Flickering reports whether the fighter is drawn through the white flash.
// A stun flashes over the opening frames of its hit animation; a flinch has no
// animation of its own, so it flickers for as long as it is held.
func (a *Actor) Flickering() bool {
	if !a.AnimWhite {
		return false
	}
	if a.State == StateAttacked {
		return a.AnimCurrentFrame < 2
	}
	return true
}

// crowding returns a living fighter on its feet whose collision box overlaps
// this one's, in front or behind.
//
// Never a teammate: a shove is something done to an enemy, and fighters on one
// team walk through one another rather than throw each other off. Bots hold
// formation and move as a crowd, which shoving them apart never let them do.
//
// Never a bot on its way round to the far side of this one either, nor the
// fighter it is walking round - see Actor.crossing.
//
// Only one on its feet: a shove is a thing done on the ground, to someone on
// it. A fighter passing through in the air - jumping up past a ledge another
// stands on, most often - is left to land first. Shoving it out of the air
// knocked a climber off the same ledge every time it tried it.
//
// The fighter the last shove landed on is left alone until the two are apart -
// while it glides clear, and for good if a wall stops it flush. Nothing else
// would stop the next tick on its feet from throwing the same shove at the same
// fighter again, and again, with no other button ever read.
func (a *Actor) crowding() *Actor {
	if a.W == nil {
		return nil
	}
	if a.PushStuck != nil && !aabb(a.collisionRect(), a.PushStuck.collisionRect()) {
		a.PushStuck = nil
	}
	for _, other := range a.W.Actors {
		if other == a || other == a.PushStuck || other.State == StateDeath || other.State == StateClash || !other.OnGround || a.allyOf(other) {
			continue
		}
		if a.crossing(other) || other.crossing(a) {
			continue
		}
		if aabb(a.collisionRect(), other.collisionRect()) {
			return other
		}
	}
	return nil
}

// pushDir is the way a shove moves other: away from the fighter throwing it.
// Two fighters on the same pixel column have no side to be on, and there the
// arms decide.
func (a *Actor) pushDir(other *Actor) int8 {
	if d := other.Pos.X - a.Pos.X; d != 0 {
		return facing(d)
	}
	return a.Orientation
}

// enterPushState turns to the crowder and starts the shove, so the arms point
// the way the crowder will go. It cannot be steered while it plays.
//
// The crowder is stunned at once, not when the arms come out. It is usually
// mid-swing - the lunge of its swings is how it walked in - and that swing
// would otherwise land during the wind-up, put the shover on guard and cut the
// shove off before it threw anyone.
func (a *Actor) enterPushState(other *Actor) {
	if a.W != nil && a.W.Authority {
		other.enterShovedState(a)
	}
	a.face(a.pushDir(other))
	a.State = StatePush
	a.PushDone = false
	a.Speed.X = 0
	a.SetAnim("resources/threefour/push.png", AnimOptions{animDuration: PushSeconds, totalFrames: PushFrames, once: true})
}

// updatePush lands the shove on the frame the arms come out, once, and ends on
// the animation. The shove is the authority's alone, as a hit is: a client is
// shown the animation, and the ground it changed hands, by the snapshot.
//
// Whoever it lands on is remembered, on a client too, until the two are apart:
// the crowder glides clear over the next few ticks, a wall can stop it short,
// and on a client nobody moves until the snapshot says so - and in none of
// those should the next tick on its feet throw the same shove again.
func (a *Actor) updatePush() {
	if a.W != nil && !a.PushDone && a.AnimCurrentFrame >= PushFrame {
		a.PushDone = true
		for _, other := range a.W.Actors {
			if other == a || other.State == StateDeath || a.allyOf(other) {
				continue
			}
			if a.crossing(other) || other.crossing(a) {
				continue
			}
			if !aabb(a.collisionRect(), other.collisionRect()) {
				continue
			}
			if a.W.Authority {
				a.shove(other)
			}
			a.PushStuck = other
		}
	}
	if a.AnimFinished {
		a.enterIdleState()
	}
}

// shove stuns a fighter and sends it gliding away from the one throwing it.
//
// The ground it gives up is a glide it is carried along, not speed it is given:
// the game moves in whole pixels, truncating every tick, so a speed that slows
// under a pixel a tick never moves anyone, and the tail of any slide is lost.
func (a *Actor) shove(other *Actor) {
	if !a.allyOf(other) {
		other.enterShovedState(a)
	}
	other.Speed.X = 0
	other.ShoveLeft, other.ShoveDir = PushDistance, a.pushDir(other)
}

// glide carries a shoved fighter one step of the way, easing out: a third of
// what is left each tick, rounded up, so 16 pixels go 6, 4, 2, 2, 1, 1. Walls
// are asked the way every other move asks them, and one that stops a step
// short ends the glide there.
func (a *Actor) glide() {
	step := (a.ShoveLeft + 2) / 3
	fromX := a.Pos.X
	a.Pos.X += float32(a.ShoveDir) * float32(step)
	a.blockSides(fromX)
	if abs(a.Pos.X-fromX) < float32(step) {
		a.ShoveLeft = 0
		return
	}
	a.ShoveLeft -= step
}

// enterShovedState is what a shove leaves the fighter it lands on: the lockout
// and the animation of a stun, so it cannot swing back through the rest of the
// shove, and none of the hit - no structure, no hit counted, no flash, no
// hit-stop. Whatever it was doing ends, as a stun ends it. One already stunned
// is left to finish the stun it has.
func (a *Actor) enterShovedState(by *Actor) {
	switch a.State {
	case StateAttacked:
		return
	case StateAttack:
		a.exitAttackState()
	case StateBlock:
		a.exitBlockState()
	}
	a.State = StateAttacked
	a.Attacker = by
	a.Speed.X = 0 // or a runner stunned mid-stride slides into the shover
	a.stopFlicker()
	a.SetAnim("resources/threefour/hit-top.png", AnimOptions{animDuration: 0.3, totalFrames: 4, once: true, reset: true})
}

// enterDeathState is where a broken structure goes. The fighter stops acting
// and stops being part of the fight; the world removes it once it has fallen.
func (a *Actor) enterDeathState(attacker *Actor) {
	a.State = StateDeath
	a.Attacker = attacker
	a.Events |= EvHit
	a.HitCount, a.HitRun, a.FlinchTicks, a.DeathTicks = 0, 0, 0, 0
	a.AnimWhite = false
	a.Speed = rl.Vector2{}
	a.AttackQueued, a.BufAttack, a.BufJump = false, false, false
	// The fallen are out of every fight at once: the place this one held at
	// whoever it was fighting, and the two at itself that its own attackers
	// hold. Waiting for the removal to do it leaves a body that stays on the
	// floor holding a place forever, and the bots behind it waiting forever
	// for a turn at the front that never comes.
	a.leaveFight()
	a.Engaged = [2]*Actor{}
	if attacker != nil {
		attacker.FreezeTicks = HitStopTicks
		attacker.restoreEnergy(MaxEnergy) // a kill puts the winner straight back in the fight
		// and some of the beating it took getting there back. The refill wait is
		// left alone: the reward is for the kill, not a reason to stop fighting.
		attacker.restoreStructure(KillStructure)
	}
	a.SetAnim("resources/threefour/death_fall.png", AnimOptions{animDuration: 0.6, totalFrames: 9, once: true})
}

// updateDeath plays the fall out, then counts the flicker off before the
// authority takes the body out of the world. A client waits to be told. A bot
// in a world where bots stay dead stops at the end of the fall, and lies there.
func (a *Actor) updateDeath() {
	if !a.AnimFinished || a.W != nil && a.W.BotsStayDead && !a.IsPlayer {
		return
	}
	a.DeathTicks++
	if a.DeathTicks >= DeathFlickerTicks && a.W != nil && a.W.Authority {
		a.W.Kill(a)
	}
}

// deathHidden reports whether the flicker is in an off span this tick.
func (a *Actor) deathHidden() bool {
	return a.State == StateDeath && a.AnimFinished && (a.DeathTicks/DeathFlickerSpan)%2 == 1
}

func (a *Actor) enterBlockState(attacker *Actor) {
	a.State = StateBlock
	a.stopFlicker()
	a.Attacker = attacker
	a.Events |= EvBlocked
	a.Speed.X = float32(attacker.Orientation) * PunchSpeed
	switch a.Updown {
	case 0:
		a.SetAnim("resources/threefour/block-low-start.png", AnimOptions{animDuration: 0.1, totalFrames: 5, once: true})
	case 1:
		a.SetAnim("resources/threefour/block-mid-start.png", AnimOptions{animDuration: 0.1, totalFrames: 5, once: true})
	case 2:
		a.SetAnim("resources/threefour/block-mid-v2-start.png", AnimOptions{animDuration: 0.1, totalFrames: 5, once: true})
	case 3:
		a.SetAnim("resources/threefour/block-all-start.png", AnimOptions{animDuration: 0.1, totalFrames: 5, once: true})
	}
}

func (a *Actor) enterIdleState() {
	a.State = StateIdle
	a.Speed.X = 0
	a.Speed.Y = 0
	if !strings.Contains(a.Anim, "updown") {
		a.SetAnimFrame("resources/threefour/updown.png", a.Updown)
	}
}

func (a *Actor) enterRunState() {
	a.State = StateRun
	a.SetAnim("resources/threefour/run.png", AnimOptions{animDuration: 0.4, totalFrames: 10})
}

func (a *Actor) updateAttacked() {
	if a.AnimFinished {
		a.exitAttackedState()
		a.enterIdleState()
	}
}

// attackBoxAt is the swing's reach at guard height updown, placed from pos.
// The simulation asks for it at the fighter's own position; the renderer asks
// for it at the interpolated one, so an outline cannot lag the sprite it marks
// and interpolation cannot reach hit resolution. Block-all (3) has no attack,
// and reports none.
func (a *Actor) attackBoxAt(updown uint8, pos rl.Vector2) (rl.Rectangle, bool) {
	return a.attackBoxFacing(updown, pos, a.Orientation)
}

// attackBoxFacing is the box a swing would cover were the fighter facing o. The
// arm covers the half of the sprite in front of it, so the two facings reach
// opposite ways from the same position. attackBoxAt is this for the way the
// fighter is facing now.
//
// The box mirrors about the middle of the collision box, not the middle of the
// 48px sprite frame. Those are not the same point - the collision box is inset
// 14 and 14 wide, so it sits 3px left of the frame's - and the box used to
// mirror about the frame while every reach test measured it against the
// collision box. The 3px went in with one sign facing right and the other
// facing left, so a swing reached 38px one way and 32px the other, and a
// fighter standing exactly on another could be hit facing left but not facing
// right. Mirroring about the box reach is measured against costs the forward
// swing 3px and gives the backward one 3px, and leaves 35 either way.
func (a *Actor) attackBoxFacing(updown uint8, pos rl.Vector2, o int8) (rl.Rectangle, bool) {
	var box rl.Rectangle
	switch updown {
	case 0:
		box = a.AttackBoxLow
	case 1:
		box = a.AttackBoxMid
	case 2:
		box = a.AttackBoxHigh
	default:
		return rl.Rectangle{}, false
	}
	mid := a.CollisionBox.X + a.CollisionBox.Width/2
	return rl.Rectangle{
		X:      pos.X + box.X + mid + (float32(o)-1)*box.Width/2,
		Y:      pos.Y + box.Y,
		Width:  box.Width,
		Height: box.Height,
	}, true
}

// swingReachFacing reports whether a swing thrown facing o would meet t: the
// box this fighter's guard height throws, carried forward by the dash the swing
// takes, against the box a hit is decided against. The bot's stopping distance
// and the shove's trigger both ask this, so neither can drift from what reaches.
func (a *Actor) swingReachFacing(t *Actor, o int8) bool {
	box, ok := a.attackBoxFacing(a.Updown, a.Pos, o)
	if !ok {
		// guarding with block-all: no swing to measure, so measure the one it
		// would throw the moment it stopped guarding
		box, _ = a.attackBoxFacing(1, a.Pos, o)
	}
	box.X += float32(o) * AttackDash
	return aabb(box, t.collisionRect())
}

func (a *Actor) updateAttack() {
	if a.Input.Attack {
		a.Events |= EvAttackStart
		a.AttackQueued = true
	}
	r1, _ := a.attackBoxAt(a.Updown, a.Pos)

	// An opening low swing answered by another in time is a clash, not a hit
	// and a block. It is looked for before the hits, or the answering fighter
	// would already be thrown into a block by the time it is asked.
	if a.W.Authority && a.openingLowSwing() {
		for _, actor := range a.W.Actors {
			if tryClash(a, actor) {
				return
			}
		}
	}

	// only the authoritative world decides hits and blocks, and one swing
	// lands on a given fighter at most once
	if a.W.Authority && !a.SwingHit {
		for _, actor := range a.W.Actors {
			if a == actor || actor.State == StateDeath || actor.State == StateAttacked || actor.State == StateClash || a.allyOf(actor) {
				continue // an ally, or a clash, is passed through: no hit, no block
			}
			r2 := actor.collisionRect()
			if !aabb(r1, r2) {
				continue
			}
			guarding := a.Updown == actor.Updown || actor.Updown == 3
			hitFrame := a.AnimCurrentFrame > attackFrames[a.Anim]
			switch {
			case actor.State == StateBlock:
				// already guarding: the outcome is settled at the hit frame
				if !hitFrame {
					continue
				}
				if guarding {
					actor.confirmBlock(a)
				} else {
					actor.exitBlockState()
					actor.takeHit(a)
				}
			case guarding:
				actor.enterBlockState(a) // provisional until the hit frame
			case hitFrame:
				actor.takeHit(a)
			}
		}
	}

	if a.AnimFinished {
		if strings.Contains(a.Anim, "back") {
			a.exitAttackState()
			a.enterIdleState()
			return
		}

		if !a.AttackQueued || !a.canSwing() {
			if a.AttackQueued { // asked for, but there is nothing left to pay with
				a.AttackQueued = false
				a.Events |= EvNoEnergy
			}
			switch a.Updown {
			case 0:
				if strings.HasSuffix(a.Anim, "attack-low-2.png") {
					a.SetAnim("resources/threefour/back-attack-low-2.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
				} else {
					a.SetAnim("resources/threefour/back-attack-low-1.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
				}
			case 1:
				a.SetAnim("resources/threefour/back-attack-mid-1.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
			case 2, 3:
				if strings.HasSuffix(a.Anim, "attack-high-2.png") {
					a.SetAnim("resources/threefour/back-attack-high-2.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
				} else {
					a.SetAnim("resources/threefour/back-attack-high-1.png", AnimOptions{animDuration: BackAttackSpeed, totalFrames: 4, once: true})
				}
			}
			return
		}

		a.AttackQueued = false
		a.SwingHit = false
		a.AttackCount++
		a.Speed.X = float32(a.Orientation) * AttackDash

		prevUpdown := a.Updown
		a.updateUpdown()
		if prevUpdown == 3 || a.Updown == 3 {
			a.enterIdleState()
			return
		}
		a.spendEnergy(EnergySwing)
		if a.Updown != prevUpdown {
			anim := fmt.Sprintf("resources/threefour/updown-atk-%d-%d.png", prevUpdown, a.Updown)
			a.SetAnim(anim, AnimOptions{animDuration: 0.4, totalFrames: 7, once: true})
			return
		}

		switch a.Updown {
		case 0:
			if a.Anim == "resources/threefour/attack-low.png" {
				a.SetAnim("resources/threefour/attack-low-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
			} else if a.Anim == "resources/threefour/attack-low-2.png" {
				a.SetAnim("resources/threefour/attack-low-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
			}
		case 1:
			a.SetAnim("resources/threefour/attack-mid-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
		case 2:
			if a.Anim == "resources/threefour/attack-high.png" {
				a.SetAnim("resources/threefour/attack-high-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
			} else if a.Anim == "resources/threefour/attack-high-2.png" {
				a.SetAnim("resources/threefour/attack-high-3.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true})
			} else if a.Anim == "resources/threefour/attack-high-3.png" {
				a.SetAnim("resources/threefour/attack-high-2.png", AnimOptions{animDuration: SecondAttackSpeed, totalFrames: 6, once: true, startFrame: 3})
			}
		}
	}
}

func (a *Actor) updateBlock() {
	// a low swing asked for from the block a low swing threw it into, in time,
	// is an answer to that swing: the two clash
	if a.Input.Attack && a.Attacker != nil && tryClash(a.Attacker, a) {
		return
	}
	if a.Attacker == nil {
		a.exitBlockState()
		a.enterIdleState()
		return
	}
	if a.Attacker.State != StateAttack && a.AnimFinished {
		a.exitBlockState()
		a.enterIdleState()
		return
	}
	if a.AnimFinished && a.Attacker.AnimCurrentFrame >= attackFrames[a.Attacker.Anim] {
		if a.W.Authority && !a.Attacker.SwingHit {
			if a.Updown != a.Attacker.Updown && a.Updown != 3 {
				attacker := a.Attacker
				a.exitBlockState()
				a.takeHit(attacker)
				return
			}
			a.confirmBlock(a.Attacker)
		}
		switch a.Anim {
		case "resources/threefour/block-all-start.png", "resources/threefour/block-all-1.png":
			a.SetAnim("resources/threefour/block-all-1.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		case "resources/threefour/block-low-start.png", "resources/threefour/block-low-1.png":
			a.SetAnim("resources/threefour/block-low-1.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		case "resources/threefour/block-mid-v2-start.png", "resources/threefour/block-mid-v2-1.png":
			a.SetAnim("resources/threefour/block-mid-v2-1.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		case "resources/threefour/block-mid-start.png", "resources/threefour/block-mid-2.png":
			a.SetAnim("resources/threefour/block-mid-1.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		case "resources/threefour/block-mid-1.png":
			a.SetAnim("resources/threefour/block-mid-2.png", AnimOptions{animDuration: BlockSpeed, totalFrames: 3, once: true, reset: true})
		}
	}
}
func (a *Actor) enterJumpState(dt float32) {
	a.OnGround = false
	a.Speed.Y = MaxJumpSpeed * dt
	a.State = StateJump
}

// enterRunJumpState commits a running fighter to its jump: the animation starts
// on the tick the input arrives, and the fighter stays on the ground until the
// frame it pushes off on. There is no separate wind-up state - the anticipation
// is the two frames the animation draws before the push-off.
//
// A run-jump that cannot be paid for does not happen: nothing is spent, the
// fighter stays where it was, and the refusal is raised for the renderer, the
// same way a swing it cannot afford is refused.
func (a *Actor) enterRunJumpState() bool {
	if !a.canRunJump() {
		a.Events |= EvNoEnergy
		return false
	}
	a.spendEnergy(EnergyRunJump)
	a.State = StateRunJump
	a.SetAnim("resources/threefour/run-jump.png", AnimOptions{animDuration: RunJumpSeconds, totalFrames: RunJumpFrames, once: true, reset: true})
	return true
}

// launchRunJump is the push-off, on the frame the animation pushes off on. It
// leaves the ground faster than a standing jump - the run-jump is the taller
// of the two - and horizontal speed is left alone, so the run's momentum is
// what carries it.
func (a *Actor) launchRunJump(dt float32) {
	a.OnGround = false
	a.Speed.Y = MaxRunJumpSpeed * dt
}

// land puts a fighter that has reached the ground back into what it was doing:
// still moving, it runs; stopped, it stands. The threshold is the one the idle
// state already runs at, so a landing cannot leave a fighter running at a speed
// it would not run at from a standstill.
//
// A landing is not a lockout: an attack pressed on the tick it lands comes out,
// rather than being swallowed by the state change.
func (a *Actor) land() {
	if abs(a.Speed.X) > RunReach {
		a.enterRunState()
	} else {
		a.enterIdleState()
	}
	if a.Input.Attack && a.Updown != 3 {
		a.enterAttackState()
	}
}

// updateFalling puts any fighter that is off the ground and going down into
// the fall. It runs before the state switch rather than inside it, so every way
// into the air reaches it: over the top of either jump, off a ledge, or carried
// off one by a knockback. A fighter that is still rising is not falling.
//
// The climb is left alone - it is on a ledge, not in the air - and so are the
// two states that are not free to act: a stun plays its own animation until it
// ends, and the dead are falling in a sense this does not mean.
func (a *Actor) updateFalling() {
	switch a.State {
	case StateFall, StateClimb, StateAttacked, StateDeath:
		return
	}
	if !a.OnGround && a.Speed.Y > 0 {
		a.enterFallState()
	}
}

// enterFallState is every way of being in the air and going down: over the top
// of either jump, off a ledge, or put there by a knockback. It leaves speed
// alone - a fall is steered like anything else airborne.
func (a *Actor) enterFallState() {
	a.State = StateFall
	a.SetAnim("resources/threefour/fall.png", AnimOptions{animDuration: 0.3, totalFrames: 3})
}

func (a *Actor) enterJumpStartState() {
	a.State = StateJumpStart
	a.SetAnim("resources/threefour/jump-start.png", AnimOptions{animDuration: JumpStartSeconds, totalFrames: JumpStartFrames, once: true, animYOffset: -48})
}

func (a *Actor) Interp(alpha float32) {
	a.InterpolatedPos.X = a.PrevPos.X*(1-alpha) + a.Pos.X*alpha
	a.InterpolatedPos.Y = a.PrevPos.Y*(1-alpha) + a.Pos.Y*alpha
}

func (a *Actor) enterClimbState() {
	a.State = StateClimb
	a.SetAnim("resources/threefour/jump-climb.png", AnimOptions{animDuration: 1.2, totalFrames: 16, once: true})
}

func (a *Actor) exitJumpState() {
	a.CanClimbRight = false
	a.CanClimbLeft = false
}

func (a *Actor) enterTurnoverState() {
	a.State = StateTurnover
	a.SetAnim("resources/threefour/turnover.png", AnimOptions{animDuration: 0.9, totalFrames: 14, once: true, reset: true})
}

func (a *Actor) exitAttackState() {
	a.AttackCount = 0
}

func (a *Actor) exitBlockState() {
	a.Attacker = nil
}

func (a *Actor) exitAttackedState() {
	a.AnimWhite = false
}

func (a *Actor) MouseOnLeft() bool {
	return a.Input.FaceLeft
}
