//go:build tinygo

// Blink an external LED. The XIAO ESP32-C3 has no user LED: the CH LED is
// the charger indicator and is not wired to a GPIO.
//
// Wiring (active high):
//
//	D0 (GPIO2) -- 330 Ω to 1 kΩ resistor -- LED anode
//	LED cathode -- GND
//
// Change ledPin if you move the LED. Avoid D9 (GPIO9): it is the BOOT strap
// pin and must be high at reset.
package main

import (
	"time"

	"machine"
)

const ledPin = machine.D0

func main() {
	led := ledPin
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for {
		led.Set(true)
		time.Sleep(500 * time.Millisecond)
		led.Set(false)
		time.Sleep(500 * time.Millisecond)
	}
}
