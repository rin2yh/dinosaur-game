//go:build tinygo && conf2025badge

package main

import "machine"

// conf2025badge carries a buzzer on GPIO1, driven by PWM0.
func newSpeaker() speaker { return newBuzzer(machine.PWM0, machine.GPIO1) }
