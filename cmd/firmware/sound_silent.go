//go:build tinygo && !conf2025badge

package main

// newSpeaker for a board with no sound wired up. Adding one: give the
// board a file behind its own build tag whose newSpeaker calls
// newBuzzer with its pins, and exclude that tag here — see
// sound_conf2025badge.go, which is the whole of it.
func newSpeaker() speaker { return silent{} }
