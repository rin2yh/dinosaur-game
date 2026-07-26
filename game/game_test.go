package game

import "testing"

// newRun creates a game and starts a run (skips the title screen).
func newRun(seed uint32) *Game {
	g := New(seed)
	g.Update(true)
	return g
}

func newPlaying(t *testing.T) *Game {
	t.Helper()
	g := newRun(12345)
	if g.Mode() != ModePlaying {
		t.Fatalf("mode = %v, want ModePlaying", g.Mode())
	}
	return g
}

func TestTitleToPlaying(t *testing.T) {
	g := New(1)
	if g.Mode() != ModeTitle {
		t.Fatalf("initial mode = %v, want ModeTitle", g.Mode())
	}
	g.Update(false)
	if g.Mode() != ModeTitle {
		t.Fatalf("mode changed without input")
	}
	g.Update(true)
	if g.Mode() != ModePlaying {
		t.Fatalf("mode = %v after jump, want ModePlaying", g.Mode())
	}
}

func TestPlayerJumpAndLand(t *testing.T) {
	g := newPlaying(t)
	groundTop := g.player.y
	g.Update(true)
	if g.player.y >= groundTop {
		t.Fatalf("player did not rise after jump: y=%v", g.player.y)
	}
	landed := false
	for i := range 120 {
		g.Update(false)
		if g.Mode() != ModePlaying {
			t.Fatalf("game over during a plain jump at frame %d", i)
		}
		if g.player.y == groundTop {
			landed = true
			break
		}
		if g.player.y < 0 {
			t.Fatalf("player flew off screen: y=%v", g.player.y)
		}
	}
	if !landed {
		t.Fatalf("player never landed; y=%v", g.player.y)
	}
}

func TestCollisionEndsGame(t *testing.T) {
	g := newPlaying(t)
	g.enemies = append(g.enemies, newEnemy(cactusSmall, playerX))
	g.Update(false)
	if g.Mode() != ModeGameOver {
		t.Fatalf("mode = %v after collision, want ModeGameOver", g.Mode())
	}
}

func TestHighBirdHitsOnlyWhenJumping(t *testing.T) {
	// Running under a high bird is safe.
	g := newPlaying(t)
	g.enemies = append(g.enemies, newEnemy(birdHigh, playerX))
	g.Update(false)
	if g.Mode() != ModePlaying {
		t.Fatalf("high bird hit a grounded player")
	}

	// Jumping into it is not.
	g = newPlaying(t)
	g.Update(true) // start rising
	for range 3 {
		g.Update(false)
	}
	g.enemies = append(g.enemies, newEnemy(birdHigh, playerX))
	g.Update(false)
	if g.Mode() != ModeGameOver {
		t.Fatalf("high bird did not hit a jumping player (y=%v)", g.player.y)
	}
}

func TestRestartAfterGameOver(t *testing.T) {
	g := newPlaying(t)
	g.enemies = append(g.enemies, newEnemy(cactusSmall, playerX))
	g.Update(false)

	// Input right after death must be ignored.
	g.Update(true)
	if g.Mode() != ModeGameOver {
		t.Fatalf("restarted during the restart delay")
	}
	for range restartDelay + 1 {
		g.Update(false)
	}
	g.Update(true)
	if g.Mode() != ModePlaying {
		t.Fatalf("mode = %v after restart, want ModePlaying", g.Mode())
	}
	if g.Score() != 0 {
		t.Fatalf("score = %d after restart, want 0", g.Score())
	}
	if len(g.enemies) != 0 {
		t.Fatalf("enemies not cleared on restart: %d", len(g.enemies))
	}
}

func TestHiScoreKept(t *testing.T) {
	g := newPlaying(t)
	g.score = 42
	g.enemies = append(g.enemies, newEnemy(cactusBig, playerX))
	g.Update(false)
	if g.HiScore() != 42 {
		t.Fatalf("hiScore = %d, want 42", g.HiScore())
	}
}

func TestEnemiesSpawnMoveAndDespawn(t *testing.T) {
	g := newPlaying(t)
	spawned := false
	for i := 0; i < 600 && g.Mode() == ModePlaying; i++ {
		g.Update(false)
		if len(g.enemies) > 0 {
			spawned = true
			for _, e := range g.enemies {
				if e.OffScreen() {
					x, _, _, _ := e.Rect()
					t.Fatalf("off-screen enemy not despawned: x=%d", x)
				}
			}
		}
	}
	if !spawned {
		t.Fatalf("no enemy spawned in 600 frames")
	}
}

func TestScoreAndSpeedIncrease(t *testing.T) {
	g := newPlaying(t)
	// Avoid dying by removing enemies each frame.
	for range scoreEvery * 10 {
		g.enemies = g.enemies[:0]
		g.Update(false)
	}
	if g.Score() == 0 {
		t.Fatalf("score did not increase")
	}
	if g.speed <= baseSpeed {
		t.Fatalf("speed = %v, want > %v", g.speed, baseSpeed)
	}
}

// clearable reports whether an enemy of the given kind can be jumped
// over (or survived by staying grounded) at the given speed with at
// least one jump timing.
func clearable(kind enemyKind, speed float64) bool {
	for delay := range 120 {
		g := newRun(1)
		g.speed = speed
		g.enemies = append(g.enemies, newEnemy(kind, ScreenWidth))
		survived := true
		for f := 0; len(g.enemies) > 0; f++ {
			g.spawnIn = ScreenWidth * 2 // suppress further spawns
			g.Update(f == delay)
			if g.Mode() == ModeGameOver {
				survived = false
				break
			}
		}
		if survived {
			return true
		}
	}
	return false
}

func TestEnemiesClearableAtUnlockSpeed(t *testing.T) {
	cases := []struct {
		name  string
		kind  enemyKind
		speed float64
	}{
		{"cactusSmall", cactusSmall, baseSpeed},
		{"cactusBig", cactusBig, bigCactusSpeed},
		{"cactusDouble", cactusDouble, doubleCactusSpeed},
		{"birdLow", birdLow, birdSpeed},
		{"birdHigh", birdHigh, birdSpeed},
	}
	for _, c := range cases {
		if !clearable(c.kind, c.speed) {
			t.Errorf("%s is impossible to clear at its unlock speed %v", c.name, c.speed)
		}
	}
}

// pairSurvives simulates two consecutive enemies spawned gap pixels
// apart with jump presses on frames d1 and d2.
func pairSurvives(k1, k2 enemyKind, speed, gap float64, d1, d2 int) bool {
	g := newRun(1)
	g.speed = speed
	g.enemies = append(g.enemies,
		newEnemy(k1, ScreenWidth), newEnemy(k2, ScreenWidth+gap))
	for f := 0; len(g.enemies) > 0; f++ {
		g.spawnIn = ScreenWidth * 4 // suppress further spawns
		g.Update(f == d1 || f == d2)
		if g.Mode() == ModeGameOver {
			return false
		}
	}
	return true
}

// pairRobust reports whether some pair of jump timings survives even
// when both presses are off by up to r frames in any combination —
// i.e. the pair is clearable with human-scale timing tolerance, not
// just frame-perfectly.
func pairRobust(k1, k2 enemyKind, speed, gap float64, r int) bool {
	for d1 := range 70 {
		for d2 := d1 + 10; d2 < d1+110; d2++ {
			ok := true
			for i := -r; i <= r && ok; i++ {
				for j := -r; j <= r && ok; j++ {
					ok = pairSurvives(k1, k2, speed, gap, d1+i, d2+j)
				}
			}
			if ok {
				return true
			}
		}
	}
	return false
}

// TestBackToBackEnemiesClearableAtMaxSpeed guards the spawn-gap
// formula: even the tightest gap at max speed must be clearable with
// at least ±5 frames (~83ms) of slack on both jumps.
// TestEnemyTableUnlocksSorted guards spawn()'s assumption that the
// unlocked kinds always form a prefix of enemyTable.
func TestEnemyTableUnlocksSorted(t *testing.T) {
	for i := 1; i < len(enemyTable); i++ {
		if enemyTable[i].unlock < enemyTable[i-1].unlock {
			t.Fatalf("enemyTable not sorted by unlock speed at index %d", i)
		}
	}
}

func TestBackToBackEnemiesClearableAtMaxSpeed(t *testing.T) {
	minGap := minSpawnGap(maxSpeed)
	pairs := []struct {
		name   string
		k1, k2 enemyKind
	}{
		{"birdLow->cactusSmall", birdLow, cactusSmall},
		{"cactusSmall->cactusSmall", cactusSmall, cactusSmall},
		{"cactusDouble->cactusSmall", cactusDouble, cactusSmall},
	}
	for _, p := range pairs {
		if !pairRobust(p.k1, p.k2, maxSpeed, minGap, 5) {
			t.Errorf("%s needs tighter than ±5 frame timing at max speed with min gap %.0fpx", p.name, minGap)
		}
	}
}

// TestEnemiesScrollEvenlyPastPlayer guards against an enemy visibly
// stalling as it scrolls past the player and off the left edge, which
// is what an int() conversion (truncation toward zero) causes at
// x == 0 — just left of playerX.
func TestEnemiesScrollEvenlyPastPlayer(t *testing.T) {
	// Every speed the game runs at moves at least a whole pixel per
	// frame, so speed is not the free variable here — the sub-pixel
	// phase an enemy happens to reach the left edge with is. Sweep the
	// phase at both ends of the speed range, for every kind.
	for kind := range len(enemyTable) {
		for _, speed := range []float64{baseSpeed, maxSpeed} {
			for phase := range 8 {
				e := newEnemy(enemyKind(kind), 20+float64(phase)/8)
				prev, _, _, _ := e.Rect()
				for !e.OffScreen() {
					e.Update(speed)
					x, _, _, _ := e.Rect()
					if step := prev - x; step < 1 {
						t.Fatalf("enemy kind %d at speed %v phase %d: stalled at x=%d",
							kind, speed, phase, x)
					}
					prev = x
				}
			}
		}
	}
}

func TestNightModeTogglesAndReverts(t *testing.T) {
	g := newPlaying(t)
	g.score = invertEvery - 1
	for range scoreEvery {
		g.enemies = g.enemies[:0]
		g.Update(false)
	}
	if !g.Night() {
		t.Fatalf("night mode not active at score %d", g.Score())
	}
	for range invertDuration {
		g.enemies = g.enemies[:0]
		g.Update(false)
	}
	if g.Night() {
		t.Fatalf("night mode did not revert after %d frames", invertDuration)
	}
}

// boundsDisplay fails the test if the game draws outside the screen.
type boundsDisplay struct {
	t *testing.T
}

func (b boundsDisplay) SetPixel(x, y int) {
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		b.t.Fatalf("SetPixel out of bounds: (%d, %d)", x, y)
	}
}

func TestDrawStaysInBounds(t *testing.T) {
	g := New(999)
	d := boundsDisplay{t: t}
	for frame := range 3000 {
		// Mash jump occasionally to cover title, play, and game over.
		g.Update(frame%37 == 0)
		g.Draw(d)
	}
}

func TestSpriteSizesMatchConstants(t *testing.T) {
	check := func(name string, s sprite, w, h int) {
		gw, gh := s.size()
		if gw != w || gh != h {
			t.Errorf("%s size = %dx%d, want %dx%d", name, gw, gh, w, h)
		}
		for i, row := range s {
			if len(row) != w {
				t.Errorf("%s row %d width = %d, want %d", name, i, len(row), w)
			}
		}
	}
	check("dinoRun1", dinoRun1, playerW, playerH)
	check("dinoRun2", dinoRun2, playerW, playerH)
	check("dinoStand", dinoStand, playerW, playerH)
	check("dinoDead", dinoDead, playerW, playerH)
	check("birdUp", birdUp, birdW, birdH)
	check("birdDown", birdDown, birdW, birdH)
	for _, s := range []sprite{cactusSmallSprite, cactusBigSprite, cactusDoubleSprite} {
		w, _ := s.size()
		for i, row := range s {
			if len(row) != w {
				t.Errorf("cactus row %d width = %d, want %d", i, len(row), w)
			}
		}
	}
}
