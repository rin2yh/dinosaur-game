package beep

import (
	"encoding/binary"
	"math"
	"testing"
)

// decode reads a rendered buffer back into mono samples, checking the
// stereo pair agrees on the way.
func decode(t *testing.T, buf []byte) []float64 {
	t.Helper()
	if len(buf)%8 != 0 {
		t.Fatalf("buffer of %d bytes is not whole stereo frames", len(buf))
	}
	out := make([]float64, 0, len(buf)/8)
	for i := 0; i < len(buf); i += 8 {
		l := math.Float32frombits(binary.LittleEndian.Uint32(buf[i:]))
		r := math.Float32frombits(binary.LittleEndian.Uint32(buf[i+4:]))
		if l != r {
			t.Fatalf("channels differ at frame %d: %v vs %v", i/8, l, r)
		}
		out = append(out, float64(l))
	}
	return out
}

// TestRenderStaysInRangeAndSilentAtTheEdges is the guard on the
// synthesis: a rendered effect has to be the length it asked for, stay
// inside the range the mixer takes, and begin and end at silence. A
// buffer that starts or stops part-way up a waveform pops, and on
// effects this short the pop is most of what a player would hear.
func TestRenderStaysInRangeAndSilentAtTheEdges(t *testing.T) {
	cases := []struct {
		name  string
		notes []Note
	}{
		{"steady tone", []Note{{Hz: 880, Ms: 60, Gain: 0.8}}},
		{"sweep", []Note{{Hz: 420, EndHz: 120, Ms: 320, Gain: 1}}},
		{"notes around a rest", []Note{
			{Hz: 880, Ms: 60, Gain: 0.8},
			{Ms: 40},
			{Hz: 1320, Ms: 120, Gain: 0.8},
		}},
	}
	for _, c := range cases {
		want := 0
		for _, n := range c.notes {
			want += samples(n.Ms)
		}
		s := decode(t, Render(c.notes))
		if len(s) != want {
			t.Errorf("%s: %d samples, want %d", c.name, len(s), want)
			continue
		}

		peak := 0.0
		for _, v := range s {
			if math.IsNaN(v) {
				t.Fatalf("%s: NaN sample", c.name)
			}
			peak = math.Max(peak, math.Abs(v))
		}
		// Full scale is 1; past it the sample clips. Well under it is
		// a sound nobody hears over the room.
		if peak > 1 || peak < 0.2 {
			t.Errorf("%s: peak amplitude %.3f, want within (0.2, 1]", c.name, peak)
		}
		// The bound is -80 dBFS, under the quietest step 16-bit audio
		// can take. The envelope leaves a tail rather than landing on
		// a hard zero, so the edges are only silent to a tolerance —
		// but they are silent by four more orders of magnitude than
		// this, which leaves room to hear a regression coming.
		if edge := math.Max(math.Abs(s[0]), math.Abs(s[len(s)-1])); edge > 1e-4 {
			t.Errorf("%s: starts or ends at %.3g, not silence", c.name, edge)
		}
	}
}

// TestRestIsSilent checks a gain-0 note renders as actual silence, so
// the gap in a two-note chime is a gap.
func TestRestIsSilent(t *testing.T) {
	for i, v := range decode(t, Render([]Note{{Ms: 40}})) {
		if v != 0 {
			t.Fatalf("rest is not silent at sample %d: %v", i, v)
		}
	}
}

// TestEnvelopeShape pins the amplitude curve: in from silence, out to
// silence, never past full scale.
func TestEnvelopeShape(t *testing.T) {
	cases := []struct {
		name string
		at   float64
		want float64
	}{
		{"starts at silence", 0, 0},
		{"peaks where the attack ends", attack, 1},
		{"ends at silence", 1, 0},
	}
	for _, c := range cases {
		if got := envelope(c.at); got != c.want {
			t.Errorf("%s: envelope(%v) = %v, want %v", c.name, c.at, got, c.want)
		}
	}
	for i := range 101 {
		if got := envelope(float64(i) / 100); got < 0 || got > 1 {
			t.Fatalf("envelope(%.2f) = %v, outside [0, 1]", float64(i)/100, got)
		}
	}
}
