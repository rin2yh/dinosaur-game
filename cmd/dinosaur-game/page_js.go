//go:build js

package main

import "syscall/js"

// current is the color the page is already showing. Draw runs every
// frame but the palette only changes when night mode flips, and a
// syscall/js round trip is far more expensive than the comparison.
var current [4]byte

// setPageBackground paints the area around the canvas. index.html sizes
// the canvas to a 2:1 box rather than the whole viewport, so the page
// background shows as a border; driving it from the same palette entry
// the canvas is cleared with is what keeps that border invisible.
func setPageBackground(c [4]byte) {
	if c == current {
		return
	}
	current = c
	js.Global().Get("document").Get("documentElement").Get("style").
		Set("background", cssHex(c))
}

// cssHex renders a palette entry as "#rrggbb". The alpha byte is
// dropped: every entry is opaque.
func cssHex(c [4]byte) string {
	const digits = "0123456789abcdef"
	hex := []byte("#000000")
	for i, v := range c[:3] {
		hex[1+i*2] = digits[v>>4]
		hex[2+i*2] = digits[v&0x0f]
	}
	return string(hex)
}
