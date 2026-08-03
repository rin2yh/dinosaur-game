package main

import (
	"github.com/hajimehoshi/ebiten/v2/audio"

	"github.com/rin2yh/dinosaur-game/game"
	"github.com/rin2yh/dinosaur-game/internal/beep"
)

// masterVolume is the one knob for how loud the game is; the gains on
// the notes below are relative to each other, not absolute.
const masterVolume = 0.4

// effects is what each of the game's sounds is played as — a slice
// rather than a map so the render order is fixed. The shapes follow
// the game: a rising blip for leaving the ground, a two-note chime up
// a fifth for a score milestone, and a long fall for the run ending.
var effects = []struct {
	sound game.Sound
	notes []beep.Note
}{
	{game.SoundJump, []beep.Note{
		{Hz: 520, EndHz: 740, Ms: 70, Gain: 0.9},
	}},
	{game.SoundPoint, []beep.Note{
		{Hz: 880, Ms: 60, Gain: 0.8},
		{Ms: 40},
		{Hz: 1320, Ms: 120, Gain: 0.8},
	}},
	{game.SoundDie, []beep.Note{
		{Hz: 420, EndHz: 120, Ms: 320, Gain: 1},
	}},
}

// speaker holds one player per effect, ready to fire. Each player owns
// its whole effect as a buffer, so playing one costs no allocation and
// no decoding.
type speaker struct {
	players []*audio.Player
}

func newSpeaker() *speaker {
	ctx := audio.NewContext(beep.SampleRate)
	s := &speaker{players: make([]*audio.Player, len(effects))}
	for i, e := range effects {
		p := ctx.NewPlayerF32FromBytes(beep.Render(e.notes))
		p.SetVolume(masterVolume)
		s.players[i] = p
	}
	return s
}

// play sounds every effect in the set. Retriggering one that is still
// sounding restarts it from the top, which is what a game this twitchy
// wants: the jump blip should follow the button rather than queue up
// behind the copy of itself the last jump started.
func (s *speaker) play(set game.Sound) {
	for i, e := range effects {
		if !set.Has(e.sound) {
			continue
		}
		p := s.players[i]
		// Rewind only fails on a source that cannot seek, and these
		// are all byte slices.
		_ = p.Rewind()
		p.Play()
	}
}
