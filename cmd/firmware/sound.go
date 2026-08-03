//go:build tinygo

package main

import "github.com/rin2yh/dinosaur-game/game"

// speaker is where this frontend hands the effects a frame triggered.
// play is called once per game frame with that frame's set, silent
// frames included, so an implementation can count frames to decide
// when to stop a note without a second method to call.
//
// Sound is outside what koebiten offers, so playing anything here
// means driving the hardware directly, and what there is to drive
// differs per board: a buzzer on GPIO1 for conf2025badge, one soldered
// on for the badges, nothing at all on a bare zero-kb02 until one is
// wired to its expansion pins. So newSpeaker belongs to the board,
// behind its build tag, and a board without one stays quiet.
type speaker interface {
	play(set game.Sound)
}

// silent is the speaker for a board with nothing to play through. It
// is also what a board falls back to when its hardware refuses to
// configure: a game that runs without sound beats one that does not
// boot.
type silent struct{}

func (silent) play(game.Sound) {}
