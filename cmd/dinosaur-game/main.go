// Command dinosaur-game is the Ebitengine frontend for the dinosaur
// game, covering both desktop and the browser via WebAssembly. All game
// logic and rendering live in the engine-agnostic game package; this
// file only wires up input and the framebuffer.
package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/rin2yh/dinosaur-game/game"
)

// Chrome dino palette: dark gray on off-white.
var (
	bg = [4]byte{0xf7, 0xf7, 0xf7, 0xff}
	fg = [4]byte{0x53, 0x53, 0x53, 0xff}
)

// frameBuffer implements game.Display over an RGBA pixel buffer.
type frameBuffer struct {
	pix []byte
	fg  [4]byte
}

func (f *frameBuffer) SetPixel(x, y int) {
	i := (y*game.ScreenWidth + x) * 4
	copy(f.pix[i:i+4], f.fg[:])
}

func (f *frameBuffer) clear(bg [4]byte) {
	// Fill the first pixel, then double the filled region each pass.
	copy(f.pix[:4], bg[:])
	for i := 4; i < len(f.pix); i *= 2 {
		copy(f.pix[i:], f.pix[:i])
	}
}

type app struct {
	g        *game.Game
	fb       frameBuffer
	touchIDs []ebiten.TouchID
}

func (a *app) Update() error {
	jump := inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		len(inpututil.AppendJustPressedTouchIDs(a.touchIDs[:0])) > 0
	a.g.Update(jump)
	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	if a.g.Night() {
		a.fb.fg = bg
		a.fb.clear(fg)
	} else {
		a.fb.fg = fg
		a.fb.clear(bg)
	}
	a.g.Draw(&a.fb)
	screen.WritePixels(a.fb.pix)
}

func (a *app) Layout(outsideWidth, outsideHeight int) (int, int) {
	return game.ScreenWidth, game.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(game.ScreenWidth*6, game.ScreenHeight*6)
	ebiten.SetWindowTitle("Dinosaur Game")
	a := &app{
		g:  game.New(uint32(time.Now().UnixNano())),
		fb: frameBuffer{pix: make([]byte, game.ScreenWidth*game.ScreenHeight*4)},
	}
	if err := ebiten.RunGame(a); err != nil {
		log.Fatal(err)
	}
}
