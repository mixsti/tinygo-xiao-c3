//go:build tinygo

// Button on D1 with an internal pull-up and a 30 ms debounce.
//
// Wiring:
//
//	D1 (GPIO3) -- momentary button -- GND
//
// Pressed reads low. Do not use D9 (GPIO9 / BOOT) for this: holding it low
// at reset puts the chip in download mode.
package main

import (
	"time"

	"machine"
)

const (
	buttonPin = machine.D1
	debounce  = 30 * time.Millisecond
)

func main() {
	btn := buttonPin
	btn.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	pressed := false
	lastRaw := false
	lastChange := time.Now()

	println("button on D1, pull-up, active low")
	for {
		raw := !btn.Get()
		if raw != lastRaw {
			lastRaw = raw
			lastChange = time.Now()
		}
		if time.Since(lastChange) >= debounce && raw != pressed {
			pressed = raw
			if pressed {
				println("pressed")
			} else {
				println("released")
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
}
