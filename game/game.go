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
	// accel 0.001 px/frame^2 on a 600px canvas): max is ~2.2x the
	// initial speed, reached after roughly two minutes.
	baseSpeed = 1.0
	maxSpeed  = 2.2
	accel     = 0.0002

	scoreEvery = 6 // frames per score point

	// Wide or tall enemies unlock only once the scroll speed is high
	// enough that a jump can physically clear them: the airborne
	// window is fixed, so the wider the enemy, the faster it has to
	// pass under the player.
	bigCactusSpeed    = 1.15
	doubleCactusSpeed = 1.4
	birdSpeed         = 1.55

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

	speed      float64
	score      int
	hiScore    int
	nightTimer int // frames of night mode left

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
	g.mode = ModePlaying
	g.frame = 0
	g.overFrame = 0
	g.dist = 0
	g.player = newPlayer()
	g.enemies = g.enemies[:0]
	g.spawnIn = ScreenWidth
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

// Update advances the game by one frame. jumpPressed reports whether
// the jump button was pressed on this frame (edge, not level).
func (g *Game) Update(jumpPressed bool) {
	g.frame++
	switch g.mode {
	case ModeTitle:
		if jumpPressed {
			g.startRun()
		}
	case ModePlaying:
		g.updatePlaying(jumpPressed)
	case ModeGameOver:
		g.overFrame++
		if jumpPressed && g.overFrame > restartDelay {
			g.startRun()
		}
	}
}

func (g *Game) updatePlaying(jumpPressed bool) {
	g.player.Update(jumpPressed)

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

// minSpawnGap is the smallest pixel gap between consecutive enemies.
// It scales with speed (like the original) so the time between
// obstacles never drops below what a full jump plus a landing needs:
// 45*speed px ≈ 45 frames at any speed.
func minSpawnGap(speed float64) float64 {
	return 45*speed + 15
}

// spawnGapJitter is the random extra gap in pixels added on top of
// minSpawnGap.
const spawnGapJitter = 70

func (g *Game) spawn() {
	kinds := 0
	for kinds < len(enemyTable) && enemyTable[kinds].unlock <= g.speed {
		kinds++
	}
	g.enemies = append(g.enemies, newEnemy(enemyKind(g.rand(kinds)), ScreenWidth))
	g.spawnIn = minSpawnGap(g.speed) + float64(g.rand(spawnGapJitter))
}

// rand returns a pseudo-random int in [0, n) using xorshift32,
// dependency-free so it runs the same everywhere including TinyGo.
func (g *Game) rand(n int) int {
	g.rng ^= g.rng << 13
	g.rng ^= g.rng >> 17
	g.rng ^= g.rng << 5
	return int(g.rng % uint32(n))
}
