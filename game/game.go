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

	// maxSpeed leaves ~46 frames to cross the screen, the reaction time
	// everything else is tuned around. Starting much slower than this
	// stretches each enemy's pass, so the jump arc covers less of it.
	baseSpeed = 1.29
	maxSpeed  = 2.8
	accel     = 0.000216

	scoreEvery = 6 // frames per score point

	// Score at which the run stops getting harder. The speed ramp tops
	// out around 1165, so the roster and the spacing carry the rest.
	difficultyPeak = 2500

	// Gap = (gapFloor + width*gapPerWidth) * speed + gapFlat, stretched
	// by a random factor. A wide enemy earns a longer breather, and
	// gapFlat — the term that does not scale with speed — buys fewer
	// frames as the run speeds up, so gaps tighten in frames until the
	// room between two enemies is a short hop rather than a full jump.
	// gapFloor keeps something behind the narrow ones: a small cactus is
	// 8px against a jump covering ~100. Calibrated to 38-54 frames
	// between arrivals at the start and 28-38 at full speed; the
	// difficulty then narrows only the stretch, never the floor.
	gapFloor      = 6.4  // frames of scroll every enemy earns
	gapPerWidth   = 1.65 // extra frames earned per px of width
	gapFlat       = 16   // px of flat breather, the term speed erodes
	gapJitterEasy = 1.5  // gap is stretched by up to this at difficulty 0
	gapJitterHard = 1.2  // ...and by up to this at difficulty 1

	// Cap on identical kinds in a row, so the run is never a metronome.
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
	g.lastKind, g.sameKinds = 0, 0
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

// Update advances the game by one frame. jumpHeld is the button level,
// not the press edge, because how long it stays down decides how high
// the jump goes; the edge is derived here so no frontend tracks it.
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
	g.speed = speedAt(g.frame)

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

// speedAt is the scroll speed on the given frame of a run.
func speedAt(frame int) float64 {
	return min(baseSpeed+accel*float64(frame), maxSpeed)
}

// difficulty is a 0-to-1 ramp over the score, which keeps climbing
// after the scroll speed has flattened out.
func (g *Game) difficulty() float64 {
	return min(float64(g.score)/difficultyPeak, 1)
}

// minSpawnGap and maxSpawnGap bracket the gap in px an enemy of the
// given width earns behind it; the stretch picks between them.
func minSpawnGap(width int, speed float64) float64 {
	return (gapFloor+float64(width)*gapPerWidth)*speed + gapFlat
}

func maxSpawnGap(width int, speed, diff float64) float64 {
	return minSpawnGap(width, speed) * lerp(gapJitterEasy, gapJitterHard, diff)
}

// minSpawnPitch is nose to nose: the enemy's own width, so the gap is
// measured from its tail, plus the gap it earned.
func minSpawnPitch(width int, speed float64) float64 {
	return float64(width) + minSpawnGap(width, speed)
}

// spawnGap rolls the actual gap. The stretch is multiplicative, so the
// spread stays proportional at every speed.
func (g *Game) spawnGap(width int) float64 {
	lo := minSpawnGap(width, g.speed)
	hi := maxSpawnGap(width, g.speed, g.difficulty())
	return lo + (hi-lo)*float64(g.rand(101))/100
}

// lerp interpolates from easy to hard over t in [0, 1].
func lerp(easy, hard, t float64) float64 {
	return easy + (hard-easy)*t
}

func (g *Game) spawn() {
	kind := g.pickKind(unlockedKinds(g.score), g.difficulty())
	e := newEnemy(kind, ScreenWidth)
	_, _, w, _ := e.Rect()
	g.enemies = append(g.enemies, e)
	g.spawnIn = float64(w) + g.spawnGap(w)

	// A drifting bird would eat a gap already decided, so it pays with a
	// head start, or, drifting back, with extra room behind it.
	if b, ok := e.(*bird); ok {
		if lead := b.drift(g.rand(2) == 0); lead > 0 {
			b.x += lead
		} else {
			g.spawnIn -= lead
		}
	}
}

// pickKind picks among the unlocked kinds. The table's tail is its
// hard end, so rolling twice and keeping the higher roll — more often
// as the difficulty rises — shifts the mix late without locking the
// easy kinds out. Repeats are capped unless the roster is too small.
func (g *Game) pickKind(kinds int, diff float64) enemyKind {
	var k enemyKind
	harder := int(diff * 1000)
	for range kinds {
		k = enemyKind(g.rand(kinds))
		if g.rand(1000) < harder {
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
