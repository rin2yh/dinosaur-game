// Package beep synthesizes the game's sound effects as raw PCM.
//
// The effects are a few blips a few hundred milliseconds long, so they
// are generated at startup rather than shipped as files: no assets to
// license or to carry into the WASM build the browser downloads whole.
//
// This is its own package, and one that imports nothing but the
// standard library, so the arithmetic below can be tested on a machine
// with no display and no sound card — importing Ebitengine anywhere in
// a package makes its tests need an X display to even start. That is
// the whole reason for the split: this is not a general audio library,
// and Render emits the one buffer shape the Ebitengine frontend's
// players take rather than a format anything else would ask for.
package beep

import (
	"encoding/binary"
	"math"
)

// SampleRate is the rate everything here is rendered at. A player's
// source has to match its audio context, so a frontend has to open its
// context at this rate.
const SampleRate = 48000

// attack is the fraction of a note spent ramping up from silence. A
// note that starts at full amplitude starts with a click.
const attack = 0.02

// Note is one segment of an effect: a sine sweeping from Hz to EndHz
// over Ms milliseconds, fading out over its own length. A Note with
// Gain 0 is a rest.
type Note struct {
	Hz, EndHz float64 // EndHz 0 means a steady tone at Hz
	Ms        int
	Gain      float64 // relative to the other notes; 1 is full scale
}

// Render renders notes to linear PCM: 32-bit float samples, little
// endian, two channels, which is the format Ebitengine's audio players
// take. Both channels carry the same signal — these are mono blips.
func Render(notes []Note) []byte {
	total := 0
	for _, n := range notes {
		total += samples(n.Ms)
	}
	buf := make([]byte, 0, total*2*4)

	for _, n := range notes {
		count := samples(n.Ms)
		phase := 0.0
		for i := range count {
			t := float64(i) / float64(count)
			hz := n.Hz
			if n.EndHz != 0 {
				hz = n.Hz + (n.EndHz-n.Hz)*t
			}
			// Step the phase rather than evaluating sin(2πft): that
			// keeps a swept frequency continuous, and a jump in phase
			// is an audible click.
			phase += 2 * math.Pi * hz / SampleRate
			v := math.Float32bits(float32(math.Sin(phase) * n.Gain * envelope(t)))
			buf = binary.LittleEndian.AppendUint32(buf, v)
			buf = binary.LittleEndian.AppendUint32(buf, v)
		}
	}
	return buf
}

// samples is how many samples ms milliseconds is at SampleRate.
func samples(ms int) int { return ms * SampleRate / 1000 }

// envelope is a note's amplitude at t in [0, 1]: a short ramp in, then
// a decay that reaches silence at t = 1. Render never evaluates it
// there — dividing by the sample count rather than one less than it
// leaves the last sample a step short of the end, which is what keeps
// two notes in a row from repeating a sample at the seam. The tail
// that step leaves is around -150 dBFS, quieter than the smallest step
// a 24-bit DAC can take, so the note still ends at silence.
//
// Squaring the decay makes it fall away quickly, so the notes read as
// blips rather than as held tones.
func envelope(t float64) float64 {
	if t < attack {
		return t / attack
	}
	d := 1 - (t-attack)/(1-attack)
	return d * d
}
