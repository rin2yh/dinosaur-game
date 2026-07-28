//go:build !js

package main

// setPageBackground does nothing outside the browser: the desktop window
// is the game screen, with no page around it to keep in step.
func setPageBackground([4]byte) {}
