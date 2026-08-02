package game

// Sound is a set of sound effects, one bit per effect. The game
// package never makes a sound itself: it has no engine to make one
// with, and the firmware target has no speaker at all. Update only
// records what happened on the frame, and a frontend that can play
// audio polls Sounds and plays whatever it recognizes.
type Sound uint8

const (
	// SoundJump fires on the frame the player leaves the ground. A
	// press the game cannot act on — mid-air, or on the title screen —
	// stays silent, so the sound always matches what is on screen.
	SoundJump Sound = 1 << iota

	// SoundPoint fires when the score crosses a multiple of pointEvery.
	SoundPoint

	// SoundDie fires on the frame a run ends.
	SoundDie
)

// pointEvery is how many points apart the milestone chime rings. The
// score ticks every scoreEvery frames, so this is one chime per ten
// seconds of running at 60 TPS.
const pointEvery = 100

// Has reports whether s contains the effect x.
func (s Sound) Has(x Sound) bool { return s&x != 0 }

// Sounds returns the effects the most recent Update triggered. Several
// can land on one frame — a jump straight into the enemy that ends the
// run rings both — which is why this is a set rather than one value.
// Each Update starts from an empty set, so a frontend stepping the
// game more than once per tick has to read this after every step or it
// will miss the earlier ones.
func (g *Game) Sounds() Sound { return g.sounds }
