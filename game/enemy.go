package game

// Enemy is anything that ends the run when the player touches it.
// New enemy types only need to implement this interface and get an
// entry in enemyTable.
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
	birdLow  // must jump over
	birdHigh // must run under (jumping hits it)
)

// enemyTable holds each kind's unlock speed and constructor, indexed
// by enemyKind and ordered by unlock speed (spawn relies on the
// unlocked kinds forming a prefix; a test guards the ordering).
var enemyTable = [...]struct {
	unlock float64
	spawn  func(x float64) Enemy
}{
	cactusSmall:  {baseSpeed, func(x float64) Enemy { return &cactus{x: x, sprite: cactusSmallSprite} }},
	cactusBig:    {bigCactusSpeed, func(x float64) Enemy { return &cactus{x: x, sprite: cactusBigSprite} }},
	cactusDouble: {doubleCactusSpeed, func(x float64) Enemy { return &cactus{x: x, sprite: cactusDoubleSprite} }},
	birdLow:      {birdSpeed, func(x float64) Enemy { return &bird{x: x} }},
	birdHigh:     {birdSpeed, func(x float64) Enemy { return &bird{x: x, high: true} }},
}

// newEnemy creates an enemy of the given kind with its left edge at x.
func newEnemy(kind enemyKind, x float64) Enemy {
	return enemyTable[kind].spawn(x)
}

// screenX converts a world x to the pixel column a sprite starts at.
// A plain int() conversion truncates toward zero, so both (-1, 0) and
// [0, 1) land on column 0 and an enemy stalls there for an extra
// frame — right beside the player, since playerX is 6. Flooring keeps
// every pixel step evenly timed all the way off the left edge.
func screenX(x float64) int {
	i := int(x)
	if x < 0 && float64(i) != x {
		i--
	}
	return i
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
)

// bird is a flying obstacle. Low birds must be jumped over; high
// birds fly at head height and must be run under.
type bird struct {
	x    float64
	high bool
}

func (b *bird) Update(speed float64) {
	b.x -= speed
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

// cactusDoubleSprite is two small cacti with a 1px gap, sharing one
// collision box.
var cactusDoubleSprite = sideBySide(cactusSmallSprite)

// sideBySide returns a sprite made of two copies of s separated by a
// 1px transparent gap.
func sideBySide(s sprite) sprite {
	out := make(sprite, len(s))
	for i, row := range s {
		out[i] = row + "." + row
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
