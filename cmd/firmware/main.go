//go:build tinygo

// Command firmware is the koebiten frontend for the dinosaur game,
// flashed onto a microcontroller with a 128x64 1-bit OLED such as
// zero-kb02. koebiten picks the board from a build tag, so supporting
// another one is a matter of adding a target under targets/ rather than
// another command. All game logic and rendering live in the
// engine-agnostic game package; this file only wires up input and the
// display.
package main

import (
	"image/color"
	"log"

	"github.com/sago35/koebiten"
	"github.com/sago35/koebiten/hardware"
	"tinygo.org/x/drivers/pixel"

	"github.com/yuuki/dinosaur-game/game"
)

// koebiten v0.5.0 clears the display buffer to all-pixels-off and games
// draw lit pixels, so daytime is white on black. That also suits an
// emissive OLED, where a full white background would be glaring.
//
// koebiten's main branch has since added whiteBGDisplay, which flips
// this convention to black ink on white; swapping dayFG and nightFG is
// the whole fix if the dependency is upgraded past v0.5.0.
var (
	dayFG   = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	nightFG = color.RGBA{A: 0xff}

	// Inverting is what night mode means, so its background is the
	// daytime foreground.
	nightBG = pixel.NewMonochrome(dayFG.R, dayFG.G, dayFG.B)
)

// oledDisplay implements game.Display by forwarding each pixel to the
// hardware display in the current foreground color.
type oledDisplay struct {
	d  koebiten.Displayer
	fg color.RGBA
}

func (o *oledDisplay) SetPixel(x, y int) {
	o.d.SetPixel(int16(x), int16(y), o.fg)
}

type app struct {
	g    *game.Game
	disp oledDisplay
	keys []koebiten.Key
}

func (a *app) Update() error {
	a.keys = koebiten.AppendJustPressedKeys(a.keys[:0])

	// koebiten ticks every 32ms (~31 TPS) while the game assumes 60
	// TPS, so step the game twice per tick to keep wall-clock speed.
	a.g.Update(len(a.keys) > 0)
	a.g.Update(false)
	return nil
}

func (a *app) Draw(_ *koebiten.Image) {
	if a.g.Night() {
		koebiten.DrawFilledRect(nil, 0, 0, game.ScreenWidth, game.ScreenHeight, nightBG)
		a.disp.fg = nightFG
	} else {
		a.disp.fg = dayFG
	}
	a.g.Draw(&a.disp)
}

func (a *app) Layout(outsideWidth, outsideHeight int) (int, int) {
	return game.ScreenWidth, game.ScreenHeight
}

func main() {
	if err := koebiten.SetHardware(hardware.Device); err != nil {
		log.Fatal(err)
	}
	koebiten.SetWindowSize(game.ScreenWidth, game.ScreenHeight)
	koebiten.SetWindowTitle("Dinosaur Game")

	a := &app{
		// Any non-zero seed does: the board has no clock, so the game
		// mixes the player's press timing in when a run starts.
		g:    game.New(1),
		disp: oledDisplay{d: hardware.Device.GetDisplay()},
		keys: make([]koebiten.Key, 0, koebiten.KeyMax+1),
	}
	if err := koebiten.RunGame(a); err != nil {
		log.Fatal(err)
	}
}
