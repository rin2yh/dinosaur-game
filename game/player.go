package game

const (
	playerX = 6
	playerW = 14
	playerH = 15

	gravity = 0.22

	// Jump velocity, with a term that leans on the current scroll speed
	// so a faster run jumps a little higher and the airborne window does
	// not shrink in distance as the ground rushes by.
	jumpVel         = -3.6
	jumpVelPerSpeed = -0.134

	// Letting go of the button mid-rise caps the rest of the climb at
	// dropVel, which is what makes a tap a short hop and a held press a
	// full jump. The cut only applies once the player has risen
	// minJumpHeight, so even the shortest hop clears a small cactus.
	dropVel       = -1.7
	minJumpHeight = 9.4

	// playerStandY is the player's top edge when standing on the ground.
	playerStandY = groundY - playerH
)

// Player is the dinosaur. Its x position is fixed; only the vertical
// axis is simulated.
type Player struct {
	y    float64 // top edge; playerStandY when on the ground
	velY float64
}

func newPlayer() Player {
	return Player{y: playerStandY}
}

func (p *Player) onGround() bool {
	return p.y >= playerStandY && p.velY >= 0
}

// Update advances the player physics by one frame. pressed is the
// press edge and held whether the button is still down; speed is the
// current scroll speed, which the jump velocity leans on.
func (p *Player) Update(pressed, held bool, speed float64) {
	if pressed && p.onGround() {
		p.velY = jumpVel + jumpVelPerSpeed*speed
	}
	// Cut the rise short once the button is up. Checking the button
	// state every frame rather than only on the release edge keeps a
	// release below minJumpHeight from silently becoming a full jump,
	// which matters on the firmware: it samples input at ~31Hz, so the
	// shortest press it can report is already two game frames long.
	if !held && p.velY < dropVel && playerStandY-p.y >= minJumpHeight {
		p.velY = dropVel
	}
	p.velY += gravity
	p.y += p.velY
	if p.y > playerStandY {
		p.y = playerStandY
		p.velY = 0
	}
}

// HitRect returns the tighter box used for collisions. It excludes
// the tail and the sprite fringe so near misses feel fair, like the
// original's per-part collision boxes.
func (p *Player) HitRect() (x, y, w, h int) {
	return playerX + 4, int(p.y) + 1, playerW - 5, playerH - 2
}

// Draw renders the player. frame drives the running animation and
// mode selects the pose (running, standing, dead).
func (p *Player) Draw(d Display, frame int, mode Mode) {
	s := dinoStand
	switch {
	case mode == ModeGameOver:
		s = dinoDead
	case !p.onGround():
		// airborne: keep the standing pose
	case mode == ModePlaying:
		if frame/6%2 == 0 {
			s = dinoRun1
		} else {
			s = dinoRun2
		}
	}
	drawSprite(d, playerX, int(p.y), s)
}

// The dinosaur, modeled after the Chrome T-Rex: boxy head with an eye
// and a mouth notch at the right, tail tip raised at the left, a stubby
// arm, and two legs.
var dinoBody = sprite{
	"........######",
	"........#.####",
	"........######",
	"........######",
	"........####..",
	"........###...",
	".#......###...",
	".##...#####...",
	".##########...",
	"..##########..",
	"...#######....",
	"...######.....",
	"....######....",
}

var (
	dinoRun1 = withLegs(dinoBody,
		"....##..##....",
		"....###.......")
	dinoRun2 = withLegs(dinoBody,
		"....##..##....",
		"........###...")
	dinoStand = withLegs(dinoBody,
		"....##..##....",
		"....###.###...")
	dinoDead = withLegs(deadBody(),
		"....##..##....",
		"....###.###...")
)

// withLegs returns body extended with the given leg rows.
func withLegs(body sprite, legs ...string) sprite {
	return append(append(sprite{}, body...), legs...)
}

// deadBody is dinoBody with a wide-open eye.
func deadBody() sprite {
	s := append(sprite{}, dinoBody...)
	s[1] = "........#..###"
	return s
}
