// Package game implements the dinosaur game logic and rendering.
// It is engine-agnostic: no Ebitengine dependency, fixed 128x64 1-bit
// screen, and all drawing goes through the Display interface so that
// the game can be ported to other runtimes (e.g. sago35/koebiten).
package game

// Screen dimensions. 128x64 matches common small OLED displays.
const (
	ScreenWidth  = 128
	ScreenHeight = 64
)

const (
	groundY = 56 // y of the ground line

	// Speed curve scaled from the Chromium original (start 6, max 13,
	// accel 0.001 px/frame^2 on a 600px canvas). maxSpeed is set so an
	// enemy crosses this 128px screen in as many frames as one crosses
	// the original's 600px canvas at its top speed, which is the real
	// measure of how much reaction time the player gets.
	baseSpeed = 1.0
	maxSpeed  = 2.8
	accel     = 0.0002

	scoreEvery = 6 // frames per score point

	// difficultyPeak is the score at which the run stops getting
	// harder. The speed ramp alone tops out at score 1500, so
	// everything else the difficulty drives — the enemy roster, how
	// tightly enemies are packed, how often the hard ones come up —
	// carries the curve the rest of the way.
	difficultyPeak = 2500

	// Spawn spacing, shaped after the original's Obstacle.getGap:
	// width*speed + minGap*gapCoefficient, stretched by a random factor
	// up to MAX_GAP_COEFFICIENT. Two properties come with that shape.
	// An obstacle earns room in proportion to its own width, so a wide
	// one is followed by a longer breather. And because only the width
	// term scales with speed, the gap measured in frames shrinks as the
	// run speeds up, which is what the variable-height jump exists to
	// answer: late on, the room between two enemies is a short hop and
	// a landing rather than a full jump.
	//
	// The speed-scaled term gets a floor the original has no need for.
	// Its narrowest obstacle is already about a jump wide relative to
	// its canvas; a small cactus here is 8px against a jump that covers
	// 98, so a purely width-proportional gap would leave almost nothing
	// behind it. The floor is a jump's worth of frames, and the width
	// only buys the extra room a wide enemy has earned.
	//
	// Every term tightens with the difficulty, which is what keeps a run
	// getting harder after the speed has capped — the original stops
	// there. The hard ends are pinned by the clearability tests.
	gapFloorEasy    = 24   // frames of scroll every enemy earns...
	gapFloorHard    = 16.5 // ...at difficulty 0 and 1
	gapPerWidthEasy = 2.1  // extra frames earned per px of width...
	gapPerWidthHard = 1.7  // ...at difficulty 0 and 1
	gapFlatEasy     = 16   // px of flat breather at difficulty 0
	gapFlatHard     = 10   // ...and at difficulty 1
	gapJitterEasy   = 1.5  // gap is stretched by up to this at difficulty 0
	gapJitterHard   = 1.2  // ...and by up to this at difficulty 1

	// The original refuses a third obstacle of the same type in a row
	// (MAX_OBSTACLE_DUPLICATION), so a run never settles into a
	// metronome of identical jumps.
	maxSameKind = 2

	// Night mode: every invertEvery points the palette inverts for
	// invertDuration frames (12s), mirroring the original.
	invertEvery    = 700
	invertDuration = 12 * 60

	restartDelay = 30 // frames to ignore input after game over
)

// Mode is the current game state.
type Mode int

const (
	ModeTitle Mode = iota
	ModePlaying
	ModeGameOver
)

// Game holds the whole game state. Create with New and drive it by
// calling Update once per frame (60 ticks per second assumed).
type Game struct {
	mode      Mode
	frame     int // frames since current run started
	overFrame int // frames since game over
	dist      float64

	player  Player
	enemies []Enemy
	spawnIn float64 // remaining px of scroll until next spawn

	lastKind  enemyKind // kind of the last enemy spawned...
	sameKinds int       // ...and how many in a row it has been

	speed      float64
	score      int
	hiScore    int
	nightTimer int // frames of night mode left

	jumpHeld bool // jump button state on the previous frame

	rng uint32
}

// New creates a game showing the title screen. seed drives enemy
// randomness; any non-zero value is fine.
func New(seed uint32) *Game {
	if seed == 0 {
		seed = 1
	}
	g := &Game{rng: seed}
	g.startRun()       // initialize all run state...
	g.mode = ModeTitle // ...but wait on the title screen
	return g
}

// startRun resets all per-run state and enters ModePlaying.
func (g *Game) startRun() {
	// Fold the frame the run starts on into the rng. Boards without a
	// clock boot into an identical state every time, so the player's
	// press timing is the only entropy a frontend there can offer;
	// mixing it in here keeps seeding out of every frontend. The frame
	// count is small, so spread it over the whole word with Knuth's
	// multiplicative constant before mixing. rng must stay non-zero or
	// xorshift32 sticks at zero forever.
	if g.rng ^= uint32(g.frame) * 2654435761; g.rng == 0 {
		g.rng = 1
	}
	g.mode = ModePlaying
	g.frame = 0
	g.overFrame = 0
	g.dist = 0
	g.player = newPlayer()
	g.enemies = g.enemies[:0]
	g.spawnIn = ScreenWidth
	g.sameKinds = 0
	g.speed = baseSpeed
	g.score = 0
	g.nightTimer = 0
}

// Mode returns the current game state.
func (g *Game) Mode() Mode { return g.mode }

// Score returns the current run's score.
func (g *Game) Score() int { return g.score }

// HiScore returns the best score across runs.
func (g *Game) HiScore() int { return g.hiScore }

// Night reports whether night mode is active; the frontend should
// render with an inverted palette while it is.
func (g *Game) Night() bool { return g.nightTimer > 0 }

// Update advances the game by one frame. jumpHeld reports whether the
// jump button is down on this frame — the level, not the edge, because
// how long it stays down is what decides how high the jump goes. The
// press edge a frontend used to pass is derived here instead.
func (g *Game) Update(jumpHeld bool) {
	pressed := jumpHeld && !g.jumpHeld
	g.jumpHeld = jumpHeld
	g.frame++
	switch g.mode {
	case ModeTitle:
		if pressed {
			g.startRun()
		}
	case ModePlaying:
		g.updatePlaying(pressed, jumpHeld)
	case ModeGameOver:
		g.overFrame++
		if pressed && g.overFrame > restartDelay {
			g.startRun()
		}
	}
}

func (g *Game) updatePlaying(pressed, held bool) {
	g.player.Update(pressed, held, g.speed)

	// Scroll, spawn, and cull enemies.
	g.dist += g.speed
	g.spawnIn -= g.speed
	if g.spawnIn <= 0 {
		g.spawn()
	}
	live := g.enemies[:0]
	for _, e := range g.enemies {
		e.Update(g.speed)
		if !e.OffScreen() {
			live = append(live, e)
		}
	}
	g.enemies = live

	// Score and difficulty.
	if g.frame%scoreEvery == 0 {
		g.score++
		if g.score%invertEvery == 0 {
			g.nightTimer = invertDuration
		}
	}
	if g.nightTimer > 0 {
		g.nightTimer--
	}
	g.speed += accel
	if g.speed > maxSpeed {
		g.speed = maxSpeed
	}

	// Collision against the player's tight hit box.
	px, py, pw, ph := g.player.HitRect()
	for _, e := range g.enemies {
		ex, ey, ew, eh := e.Rect()
		if px < ex+ew && ex < px+pw && py < ey+eh && ey < py+ph {
			g.gameOver()
			return
		}
	}
}

func (g *Game) gameOver() {
	g.mode = ModeGameOver
	g.overFrame = 0
	if g.score > g.hiScore {
		g.hiScore = g.score
	}
}

// difficulty reports how far into the difficulty curve the run is, as
// a 0-to-1 ramp over the score. Score is the run's clock, so this keeps
// climbing long after the scroll speed has flattened out.
func (g *Game) difficulty() float64 {
	d := float64(g.score) / difficultyPeak
	if d > 1 {
		d = 1
	}
	return d
}

// minSpawnGap is the smallest gap in pixels an enemy of the given
// width earns behind it, before the random stretch. Wider enemies earn
// more, and only that term scales with speed, so the gap shrinks in
// frames as the run speeds up. The flat term is the breather, and it
// is what the difficulty keeps taking away after the speed has capped.
func minSpawnGap(width int, speed, diff float64) float64 {
	frames := lerp(gapFloorEasy, gapFloorHard, diff) +
		float64(width)*lerp(gapPerWidthEasy, gapPerWidthHard, diff)
	return frames*speed + lerp(gapFlatEasy, gapFlatHard, diff)
}

// spawnGap is minSpawnGap stretched by a random factor, the original's
// multiplicative jitter: the spread stays proportional at every speed
// instead of mattering less and less as the gaps grow.
func (g *Game) spawnGap(width int, diff float64) float64 {
	stretch := lerp(gapJitterEasy, gapJitterHard, diff)
	return minSpawnGap(width, g.speed, diff) * (1 + float64(g.rand(101))/100*(stretch-1))
}

// lerp interpolates from easy to hard over t in [0, 1].
func lerp(easy, hard, t float64) float64 {
	return easy + (hard-easy)*t
}

func (g *Game) spawn() {
	diff := g.difficulty()
	kinds := 0
	for kinds < len(enemyTable) && enemyTable[kinds].unlock <= g.score {
		kinds++
	}
	kind := g.pickKind(kinds, diff)
	e := newEnemy(kind, ScreenWidth)
	_, _, w, _ := e.Rect()

	// The gap is measured from this enemy's tail to the next one's
	// nose, as in the original, so an enemy's own width never eats into
	// the room the player was promised.
	g.enemies = append(g.enemies, e)
	g.spawnIn = float64(w) + g.spawnGap(w, diff)

	// A bird that drifts off the scroll speed would otherwise close on
	// whatever is ahead of it, or be closed on by whatever follows, in
	// both cases eating a gap that was already decided. Push the bird
	// forward, or the next spawn back, by exactly what the drift is
	// worth over the run in to the player.
	if b, ok := e.(*bird); ok {
		b.offset = birdSpeedOffset
		if g.rand(2) == 0 {
			b.offset = -birdSpeedOffset
		}
		if lead := (ScreenWidth - playerX) * b.offset; lead > 0 {
			b.x += lead
		} else {
			g.spawnIn -= lead
		}
	}
}

// pickKind picks one of the first kinds entries of enemyTable. The
// table is ordered by unlock score, so its tail is the hard end of the
// roster: rolling twice and keeping the higher roll — with a
// probability that grows with the difficulty — shifts the mix towards
// the newly unlocked kinds late in a run without ever locking the easy
// ones out. Runs of the same kind are capped the way the original caps
// them, except when the roster is too small to offer an alternative.
func (g *Game) pickKind(kinds int, diff float64) enemyKind {
	k := enemyKind(0)
	for try := 0; try < kinds; try++ {
		k = enemyKind(g.rand(kinds))
		if g.rand(1000) < int(diff*1000) {
			if k2 := enemyKind(g.rand(kinds)); k2 > k {
				k = k2
			}
		}
		if k != g.lastKind || g.sameKinds < maxSameKind {
			break
		}
	}
	if k == g.lastKind {
		g.sameKinds++
	} else {
		g.lastKind, g.sameKinds = k, 1
	}
	return k
}

// rand returns a pseudo-random int in [0, n) using xorshift32,
// dependency-free so it runs the same everywhere including TinyGo.
func (g *Game) rand(n int) int {
	g.rng ^= g.rng << 13
	g.rng ^= g.rng >> 17
	g.rng ^= g.rng << 5
	return int(g.rng % uint32(n))
}
