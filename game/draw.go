package game

// Display is the only interface the game draws through. A frontend
// implements it over its own framebuffer, which keeps rendering
// portable: Ebitengine writes RGBA pixels, a koebiten port can map it
// to a monochrome OLED directly.
type Display interface {
	// SetPixel turns the pixel at (x, y) on (foreground color).
	// Coordinates are guaranteed to be inside the screen.
	SetPixel(x, y int)
}

// sprite is a 1-bit bitmap; '#' is an opaque pixel, anything else is
// transparent. All rows must have equal length.
type sprite []string

func (s sprite) size() (w, h int) {
	if len(s) == 0 {
		return 0, 0
	}
	return len(s[0]), len(s)
}

// Draw renders the current frame onto d. The frontend is expected to
// clear the framebuffer to the background color beforehand.
func (g *Game) Draw(d Display) {
	g.drawGround(d)
	for _, e := range g.enemies {
		e.Draw(d, g.frame)
	}
	g.player.Draw(d, g.frame, g.mode)
	g.drawHUD(d)
}

func (g *Game) drawGround(d Display) {
	// Ground line plus scrolling bumps below it for a sense of speed.
	off := int(g.dist)
	for x := range ScreenWidth {
		d.SetPixel(x, groundY)
		wx := x + off
		if wx%13 == 0 {
			d.SetPixel(x, groundY+3)
		}
		if wx%29 == 5 {
			d.SetPixel(x, groundY+5)
		}
	}
}

const (
	scoreDigits = 5
	scoreMax    = 99999 // largest value scoreDigits digits can show
	scoreW      = scoreDigits * glyphW
)

func (g *Game) drawHUD(d Display) {
	scoreX := ScreenWidth - 2 - scoreW
	drawNumber(d, scoreX, 2, g.score)
	if g.hiScore > 0 {
		hiX := scoreX - 4 - scoreW - textWidth("HI ")
		drawText(d, hiX, 2, "HI ")
		drawNumber(d, hiX+textWidth("HI "), 2, g.hiScore)
	}
	switch g.mode {
	case ModeTitle:
		if blink(g.frame) {
			drawTextCentered(d, 24, "PRESS JUMP")
		}
	case ModeGameOver:
		drawTextCentered(d, 20, "GAME OVER")
		if g.overFrame > restartDelay && blink(g.overFrame) {
			drawTextCentered(d, 32, "PRESS JUMP")
		}
	}
}

// blink reports the "on" phase of a half-second blink for the given
// frame counter.
func blink(frame int) bool {
	return frame/30%2 == 0
}

// drawNumber draws n as scoreDigits zero-padded digits without
// allocating (the render path runs 60x per second).
func drawNumber(d Display, x, y, n int) {
	if n > scoreMax {
		n = scoreMax
	}
	for i := scoreDigits - 1; i >= 0; i-- {
		drawGlyph(d, x+i*glyphW, y, rune('0'+n%10))
		n /= 10
	}
}

func drawSprite(d Display, x, y int, s sprite) {
	w, h := s.size()
	for sy := range h {
		py := y + sy
		if py < 0 || py >= ScreenHeight {
			continue
		}
		row := s[sy]
		for sx := range w {
			px := x + sx
			if px < 0 || px >= ScreenWidth {
				continue
			}
			if row[sx] == '#' {
				d.SetPixel(px, py)
			}
		}
	}
}

func drawText(d Display, x, y int, s string) {
	for _, r := range s {
		drawGlyph(d, x, y, r)
		x += glyphW
	}
}

func drawGlyph(d Display, x, y int, r rune) {
	if g, ok := font[r]; ok {
		drawSprite(d, x, y, sprite(g[:]))
	}
}

func drawTextCentered(d Display, y int, s string) {
	drawText(d, (ScreenWidth-textWidth(s))/2, y, s)
}
