//go:build tinygo

package main

import (
	"machine"

	"tinygo.org/x/drivers/tone"

	"github.com/rin2yh/dinosaur-game/game"
)

// A buzzer is one square wave with no volume control, so the effects
// the Ebitengine frontend plays survive here only as pitch and length:
// a rising pair for a jump, a fifth up for the milestone, and a fall
// for the end of a run. Nothing below is board-specific — a board
// brings its pins and calls newBuzzer.

// note is one step of an effect: a MIDI note held for frames game
// frames, or, with note 0 — below A0, so never a real one — that many
// frames of silence. Frames are game frames, two to a koebiten tick,
// so about 16ms each.
//
// frames is a uint8 to keep the whole thing two bytes. TinyGo gives a
// package-level slice a writable copy in RAM whether or not anything
// ever writes to it, so the width is the difference between the tables
// costing 28 bytes of it and 112.
type note struct {
	note   tone.Note
	frames uint8
}

// Every effect ends in a rest, which is what silences the buzzer when
// it runs out. Without it the last note would ring on, and stopping a
// frame early instead would cut a two-frame note down to one.
var (
	jumpNotes = []note{
		{tone.C5, 2},  // 523Hz
		{tone.FS5, 2}, // 740Hz
		{0, 1},
	}
	pointNotes = []note{
		{tone.A5, 4}, // 880Hz
		{0, 2},
		{tone.E6, 7}, // 1319Hz
		{0, 1},
	}
	dieNotes = []note{
		{tone.GS4, 3}, // 415Hz
		{tone.E4, 3},
		{tone.C4, 3},
		{tone.GS3, 3},
		{tone.E3, 3},
		{tone.B2, 3}, // 123Hz
		{0, 1},
	}
)

// buzzer plays one effect at a time, stepping through its notes a
// frame at a time.
type buzzer struct {
	speaker tone.Speaker
	notes   []note // what is left of the effect playing, if any...
	elapsed uint8  // ...and how many frames its first note has had
}

// newBuzzer returns a speaker driving a buzzer on the given pin, or a
// silent one if the hardware refuses to configure: a game that runs
// without sound beats one that does not boot.
func newBuzzer(pwm tone.PWM, pin machine.Pin) speaker {
	s, err := tone.New(pwm, pin)
	if err != nil {
		return silent{}
	}
	return &buzzer{speaker: s}
}

// play advances the effect by a frame, starting a new one if the set
// calls for it. A buzzer is one voice, so a new effect cuts off
// whatever is sounding rather than waiting its turn — and when two
// land on the same frame the more urgent one wins, which is why a run
// ends on its own note rather than on the jump that ended it.
func (b *buzzer) play(set game.Sound) {
	switch {
	case set.Has(game.SoundDie):
		b.start(dieNotes)
	case set.Has(game.SoundPoint):
		b.start(pointNotes)
	case set.Has(game.SoundJump):
		b.start(jumpNotes)
	}

	if len(b.notes) == 0 {
		return
	}
	n := b.notes[0]
	if b.elapsed == 0 {
		if n.note == 0 {
			b.speaker.Stop()
		} else {
			b.speaker.SetNote(n.note)
		}
	}
	if b.elapsed++; b.elapsed == n.frames {
		b.notes, b.elapsed = b.notes[1:], 0
	}
}

func (b *buzzer) start(notes []note) {
	b.notes, b.elapsed = notes, 0
}
