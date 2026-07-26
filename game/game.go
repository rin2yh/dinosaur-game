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

	// Spawn spacing. The gapFrames values are frames of scroll, so the
	// spacing they set means the same thing at any speed; the jitter
	// and the pad are flat pixels. Easy values apply at difficulty 0
	// and hard ones at difficulty 1, so obstacles come both closer
	// together and less spread out as a run goes on.
	gapFramesEasy = 45
	gapFramesHard = 40
	gapJitterEasy = 70 // px of random extra gap
	gapJitterHard = 25
	gapPad        = 15 // px of flat extra gap at every difficulty

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

// minSpawnGap is the smallest pixel gap between consecutive enemies.
// It scales with speed (like the original) so the gap is a constant
// number of frames at any speed, never less than what a full jump plus
// a landing needs, and it tightens as the difficulty rises.
func minSpawnGap(speed, diff float64) float64 {
	return lerp(gapFramesEasy, gapFramesHard, diff)*speed + gapPad
}

// spawnGapJitter is the random extra gap in pixels added on top of
// minSpawnGap. It shrinks with difficulty, so late in a run enemies
// come at close to the minimum spacing instead of being spread out by
// the luck of the draw.
func spawnGapJitter(diff float64) int {
	return int(lerp(gapJitterEasy, gapJitterHard, diff))
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
	g.enemies = append(g.enemies, newEnemy(kind, spawnX(kind)))
	g.spawnIn = minSpawnGap(g.speed, diff) + float64(g.rand(spawnGapJitter(diff)))
}

// pickKind picks one of the first kinds entries of enemyTable. The
// table is ordered by unlock score, so its tail is the hard end of the
// roster: rolling twice and keeping the higher roll — with a
// probability that grows with the difficulty — shifts the mix towards
// the newly unlocked kinds late in a run without ever locking the easy
// ones out.
func (g *Game) pickKind(kinds int, diff float64) enemyKind {
	k := g.rand(kinds)
	if g.rand(1000) < int(diff*1000) {
		if k2 := g.rand(kinds); k2 > k {
			k = k2
		}
	}
	return enemyKind(k)
}

// rand returns a pseudo-random int in [0, n) using xorshift32,
// dependency-free so it runs the same everywhere including TinyGo.
func (g *Game) rand(n int) int {
	g.rng ^= g.rng << 13
	g.rng ^= g.rng >> 17
	g.rng ^= g.rng << 5
	return int(g.rng % uint32(n))
}
