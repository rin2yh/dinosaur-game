package game

import "testing"

// newRun creates a game and starts a run (skips the title screen).
// The button is released again on the way out, so the next press a
// test makes is an edge the game will act on.
func newRun(seed uint32) *Game {
	g := New(seed)
	g.Update(true)
	g.Update(false)
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

// TestJumpHeightFollowsHold is the guard on the ported variable-height
// jump: how long the button stays down has to decide how high and how
// long the player flies, or the tight late-game gaps have no answer.
func TestJumpHeightFollowsHold(t *testing.T) {
	fly := func(hold int) (peak float64, frames int) {
		g := newRun(1)
		peak = playerStandY
		for f := range 200 {
			g.enemies = g.enemies[:0]
			g.spawnIn = ScreenWidth * 4
			g.Update(press(f, 0, hold))
			peak = min(peak, g.player.y)
			if !g.player.onGround() {
				frames++
			} else if frames > 0 {
				break
			}
		}
		return playerStandY - peak, frames
	}

	tapH, tapF := fly(2)
	holdH, holdF := fly(40)
	if tapH < minJumpHeight {
		t.Errorf("the shortest press rises %.1fpx, below the %.1fpx floor", tapH, minJumpHeight)
	}
	if tapH >= holdH || tapF >= holdF {
		t.Errorf("holding did not buy a bigger jump: tap %.1fpx/%df, hold %.1fpx/%df",
			tapH, tapF, holdH, holdF)
	}
	// A tap has to stay useful, or the cut is a trap rather than a tool.
	if _, _, _, h := newEnemy(cactusSmall, 0).Rect(); tapH < float64(h) {
		t.Errorf("the shortest press rises %.1fpx, not enough for a %dpx cactus", tapH, h)
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

// holds are the button hold lengths the timing searches try: the
// shortest press the firmware can express, a human tap, and a hold
// long enough to never be cut short.
var holds = []int{2, 6, 12, 40}

// press reports whether the button is down on frame f for a press that
// starts on frame start and is held for hold frames.
func press(f, start, hold int) bool {
	return f >= start && f < start+hold
}

// clearable reports whether an enemy of the given kind can be jumped
// over (or survived by staying grounded) at the given speed with at
// least one jump timing.
func clearable(kind enemyKind, speed float64) bool {
	for _, hold := range holds {
		for delay := range 120 {
			g := newRun(1)
			g.speed = speed
			g.enemies = append(g.enemies, newEnemy(kind, ScreenWidth))
			survived := true
			for f := 0; len(g.enemies) > 0; f++ {
				g.spawnIn = ScreenWidth * 2 // suppress further spawns
				g.Update(press(f, delay, hold))
				if g.Mode() == ModeGameOver {
					survived = false
					break
				}
			}
			if survived {
				return true
			}
		}
	}
	return false
}

// speedAtScore returns the scroll speed a run has at the given score.
// The speed ramp is driven by frames and the score is just a frame
// counter divided by scoreEvery, so the two are locked together.
func speedAtScore(score int) float64 {
	speed := baseSpeed + accel*scoreEvery*float64(score)
	return min(speed, maxSpeed)
}

// TestEnemiesClearableAtUnlockScore is the fairness guard on the
// roster: an enemy that unlocks before the scroll is fast enough to
// carry it under the player within one jump is impossible, not hard.
func TestEnemiesClearableAtUnlockScore(t *testing.T) {
	for kind := range len(enemyTable) {
		speed := speedAtScore(enemyTable[kind].unlock)
		if !clearable(enemyKind(kind), speed) {
			t.Errorf("enemy kind %d is impossible to clear at speed %v, the speed at its unlock score %d",
				kind, speed, enemyTable[kind].unlock)
		}
	}
}

// pairSurvives simulates two consecutive enemies spawned gap pixels
// apart, with the button pressed on frames d1 and d2 and held for h1
// and h2 frames. d2 <= d1 means a single press covers both, which is
// how the tightest gaps have to be cleared.
func pairSurvives(k1, k2 enemyKind, speed, gap float64, d1, h1, d2, h2 int) bool {
	g := newRun(1)
	g.speed = speed
	g.enemies = append(g.enemies,
		newEnemy(k1, ScreenWidth), newEnemy(k2, ScreenWidth+gap))
	for f := 0; len(g.enemies) > 0; f++ {
		g.spawnIn = ScreenWidth * 4 // suppress further spawns
		g.Update(press(f, d1, h1) || press(f, d2, h2))
		if g.Mode() == ModeGameOver {
			return false
		}
	}
	return true
}

// pairRobust reports whether some pair of jump timings survives even
// when both presses are off by up to r frames in any combination —
// i.e. the pair is clearable with human-scale timing tolerance, not
// just frame-perfectly. Clearing both with one press counts, and is
// the only answer once the gap is shorter than a jump.
func pairRobust(k1, k2 enemyKind, speed, gap float64, r int) bool {
	// The second press has to be searched all the way out to when the
	// second enemy actually arrives, which at low speed is far later
	// than any fixed window would reach.
	span := int(gap/speed) + 60
	for _, h1 := range holds {
		for _, h2 := range holds {
			for d1 := range 70 {
				for d2 := d1; d2 < d1+span; d2++ {
					ok := true
					for i := -r; i <= r && ok; i++ {
						for j := -r; j <= r && ok; j++ {
							ok = pairSurvives(k1, k2, speed, gap, d1+i, h1, d2+j, h2)
						}
					}
					if ok {
						return true
					}
				}
			}
		}
	}
	return false
}

// TestEnemyTableUnlocksSorted guards spawn()'s assumption that the
// unlocked kinds always form a prefix of enemyTable.
func TestEnemyTableUnlocksSorted(t *testing.T) {
	for i := 1; i < len(enemyTable); i++ {
		if enemyTable[i].unlock < enemyTable[i-1].unlock {
			t.Fatalf("enemyTable not sorted by unlock score at index %d", i)
		}
	}
}

// pairSlack is the timing slop the tightest spawn gap has to leave on
// every press. The original leaves about this much at its own tightest
// draw, and this game is calibrated to sit alongside it: asking for
// more is what made an earlier cut of this noticeably gentler than the
// original, and asking for less would be trusting frame-perfect play.
const pairSlack = 3

// TestBackToBackEnemiesClearable guards the spawn-gap formula at its
// tightest — the minimum gap with no random stretch — at both ends of
// the speed ramp.
func TestBackToBackEnemiesClearable(t *testing.T) {
	pairs := []struct {
		name   string
		k1, k2 enemyKind
	}{
		{"birdLow->cactusSmall", birdLow, cactusSmall},
		{"cactusSmall->cactusSmall", cactusSmall, cactusSmall},
		{"cactusSmall->cactusBig", cactusSmall, cactusBig},
		{"cactusBig->cactusSmall", cactusBig, cactusSmall},
		{"cactusDouble->cactusSmall", cactusDouble, cactusSmall},
		{"cactusTriple->cactusSmall", cactusTriple, cactusSmall},
		{"cactusSmall->cactusTriple", cactusSmall, cactusTriple},
		{"cactusSmall->birdLow", cactusSmall, birdLow},
	}
	for _, p := range pairs {
		// The slowest the run can be when this pair is possible at all
		// is the speed at the later of the two unlock scores, and the
		// gap is at its tightest in frames at max speed. Check both.
		unlock := max(enemyTable[p.k1].unlock, enemyTable[p.k2].unlock)
		for _, speed := range []float64{speedAtScore(unlock), maxSpeed} {
			// The gap the leading enemy earns, measured tail to nose
			// the way spawn() lays it out, with no random stretch.
			_, _, w, _ := newEnemy(p.k1, 0).Rect()
			gap := float64(w) + minSpawnGap(w, speed)
			if !pairRobust(p.k1, p.k2, speed, gap, pairSlack) {
				t.Errorf("%s at speed %.2f needs tighter than ±%d frame timing with min gap %.0fpx",
					p.name, speed, pairSlack, gap)
			}
		}
	}
}

// TestDifficultyKeepsRisingPastMaxSpeed is the regression guard for
// the complaint this scaling exists to fix: once the speed ramp
// flattens out, the run has to keep getting harder some other way.
func TestDifficultyKeepsRisingPastMaxSpeed(t *testing.T) {
	flat := 0
	for score := 0; speedAtScore(score) >= maxSpeed; score++ {
		flat = score
	}
	if flat >= difficultyPeak {
		t.Fatalf("speed maxes out at score %d, at or past difficultyPeak %d: nothing scales after that",
			flat, difficultyPeak)
	}

	// Between the speed cap and the difficulty peak, enemies must
	// still be arriving measurably closer together — both the floor
	// and the widest the random stretch can make it.
	const w = 8 // a small cactus, the most common enemy
	worst := func(diff float64) float64 {
		return minSpawnGap(w, maxSpeed) * lerp(gapJitterEasy, gapJitterHard, diff)
	}
	early, late := worst(float64(flat)/difficultyPeak), worst(1)
	if late >= early {
		t.Errorf("worst-case spawn gap did not tighten after the speed cap: %.0fpx -> %.0fpx", early, late)
	}

	// ...and the roster must still be growing, or the last stretch is
	// the same handful of enemies at the same spacing.
	unlockedAt := func(score int) int {
		n := 0
		for n < len(enemyTable) && enemyTable[n].unlock <= score {
			n++
		}
		return n
	}
	if unlockedAt(difficultyPeak) <= unlockedAt(flat/2) {
		t.Errorf("no new enemy kind unlocks between score %d and %d", flat/2, difficultyPeak)
	}
}

// TestHarderEnemiesGetMoreCommon checks the late-game mix actually
// shifts: the same roster picked uniformly forever would keep the hard
// kinds as rare as the day they unlocked.
func TestHarderEnemiesGetMoreCommon(t *testing.T) {
	const kinds = len(enemyTable)
	count := func(diff float64) int {
		g := newRun(7)
		hard := 0
		for range 4000 {
			if int(g.pickKind(kinds, diff)) >= kinds/2 {
				hard++
			}
		}
		return hard
	}
	easy, late := count(0), count(1)
	if late <= easy {
		t.Errorf("hard-half picks per 4000 spawns: %d at difficulty 0, %d at difficulty 1; want more when harder",
			easy, late)
	}
}

// TestDifficultyRampIsMonotonic checks the curve only ever goes up and
// then holds, so a longer run is never an easier one.
func TestDifficultyRampIsMonotonic(t *testing.T) {
	g := newRun(1)
	prev := -1.0
	for score := 0; score <= difficultyPeak*2; score += 25 {
		g.score = score
		d := g.difficulty()
		if d < prev {
			t.Fatalf("difficulty dropped at score %d: %v -> %v", score, prev, d)
		}
		if d > 1 {
			t.Fatalf("difficulty exceeded 1 at score %d: %v", score, d)
		}
		prev = d
	}
	if prev != 1 {
		t.Fatalf("difficulty never reached 1; got %v", prev)
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
