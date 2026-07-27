package game

// Enemy is anything that ends the run when the player touches it. A
// new type needs this interface and an entry in enemyTable; only one
// that moves off the scroll speed, as bird does, concerns spawn().
type Enemy interface {
	// Update advances the enemy by one frame at the given scroll
	// speed in pixels per frame.
	Update(speed float64)
	// Rect returns the collision box.
	Rect() (x, y, w, h int)
	// Draw renders the enemy; frame drives animation.
	Draw(d Display, frame int)
	// OffScreen reports whether the enemy has fully scrolled out.
	OffScreen() bool
}

type enemyKind int

const (
	cactusSmall enemyKind = iota
	cactusBig
	cactusDouble
	birdLow      // must jump over
	birdHigh     // must run under (jumping hits it)
	cactusTriple // widest ground obstacle
)

// enemyTable holds each kind's unlock score and constructor, indexed
// by enemyKind and ordered by unlock score (spawn relies on the
// unlocked kinds forming a prefix, and picks harder kinds more often
// from the tail; tests guard both the ordering and that every kind is
// clearable at the speed the run has when it unlocks).
var enemyTable = [...]struct {
	unlock int
	spawn  func(x float64) Enemy
}{
	cactusSmall:  {0, func(x float64) Enemy { return &cactus{x: x, sprite: cactusSmallSprite} }},
	cactusBig:    {125, func(x float64) Enemy { return &cactus{x: x, sprite: cactusBigSprite} }},
	cactusDouble: {330, func(x float64) Enemy { return &cactus{x: x, sprite: cactusDoubleSprite} }},
	birdLow:      {450, func(x float64) Enemy { return &bird{x: x} }},
	birdHigh:     {450, func(x float64) Enemy { return &bird{x: x, high: true} }},
	cactusTriple: {800, func(x float64) Enemy { return &cactus{x: x, sprite: cactusTripleSprite} }},
}

// newEnemy creates an enemy of the given kind with its left edge at x.
func newEnemy(kind enemyKind, x float64) Enemy {
	return enemyTable[kind].spawn(x)
}

// unlockedKinds returns how many kinds have unlocked by the given
// score. They are always the first n entries; a test guards that.
func unlockedKinds(score int) int {
	n := 0
	for n < len(enemyTable) && enemyTable[n].unlock <= score {
		n++
	}
	return n
}

// cactus is a ground obstacle; the variants differ only in sprite.
type cactus struct {
	x      float64
	sprite sprite
}

func (c *cactus) Update(speed float64) {
	c.x -= speed
}

func (c *cactus) Rect() (x, y, w, h int) {
	w, h = c.sprite.size()
	return screenX(c.x), groundY - h, w, h
}

func (c *cactus) Draw(d Display, _ int) {
	x, y, _, _ := c.Rect()
	drawSprite(d, x, y, c.sprite)
}

func (c *cactus) OffScreen() bool {
	_, _, w, _ := c.Rect()
	return c.x+float64(w) <= 0
}

const (
	birdW         = 10
	birdH         = 6
	birdFlyHeight = 4 // gap between a low bird and the ground

	// Every bird flies off the scroll speed, half faster and half
	// slower: a fast one gives the least warning in the game. A fraction
	// rather than a flat px/frame, so it holds across the speed range.
	birdSpeedOffset = 0.12
)

// bird is a flying obstacle. Low birds must be jumped over; high
// birds fly at head height and must be run under. offset is the
// fraction of the scroll speed this bird flies faster (or, negative,
// slower) than the ground.
type bird struct {
	x      float64
	high   bool
	offset float64
}

// drift sets the bird's speed offset and returns the head start it
// needs for the gap ahead of it to survive. A slow bird returns a
// negative lead: the enemy behind it needs that much extra room.
func (b *bird) drift(fast bool) (lead float64) {
	b.offset = birdSpeedOffset
	if !fast {
		b.offset = -birdSpeedOffset
	}
	return (ScreenWidth - playerX) * b.offset
}

func (b *bird) Update(speed float64) {
	b.x -= speed * (1 + b.offset)
}

func (b *bird) Rect() (x, y, w, h int) {
	y = groundY - birdH - birdFlyHeight
	if b.high {
		y = groundY - playerH - birdH - birdFlyHeight
	}
	return screenX(b.x), y, birdW, birdH
}

func (b *bird) Draw(d Display, frame int) {
	x, y, _, _ := b.Rect()
	if frame/8%2 == 0 {
		drawSprite(d, x, y, birdUp)
	} else {
		drawSprite(d, x, y, birdDown)
	}
}

func (b *bird) OffScreen() bool {
	_, _, w, _ := b.Rect()
	return b.x+float64(w) <= 0
}

var cactusSmallSprite = sprite{
	"...##...",
	"...##...",
	"#..##..#",
	"#..##..#",
	"#..##..#",
	"#..##..#",
	"########",
	"...##...",
	"...##...",
	"...##...",
	"...##...",
	"...##...",
}

var cactusBigSprite = sprite{
	"....##....",
	"....##....",
	"....##....",
	"#...##...#",
	"#...##...#",
	"#...##...#",
	"#...##...#",
	"#...##...#",
	"##..##..##",
	".########.",
	"....##....",
	"....##....",
	"....##....",
	"....##....",
	"....##....",
	"....##....",
}

// The multi-cactus variants are small cacti in a row with 1px gaps,
// each sharing one collision box.
var (
	cactusDoubleSprite = sideBySide(cactusSmallSprite, 2)
	cactusTripleSprite = sideBySide(cactusSmallSprite, 3)
)

// sideBySide returns a sprite made of n copies of s, each separated by
// a 1px transparent gap.
func sideBySide(s sprite, n int) sprite {
	out := make(sprite, len(s))
	for i, row := range s {
		next := "." + row
		out[i] = row
		for range n - 1 {
			out[i] += next
		}
	}
	return out
}

var birdUp = sprite{
	"......#...",
	"......##..",
	"......###.",
	"#########.",
	".#########",
	"...####...",
}

var birdDown = sprite{
	".#########",
	"#########.",
	"...####...",
	"......###.",
	"......##..",
	"......#...",
}
