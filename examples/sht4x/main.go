//go:build tinygo

// Read a Sensirion SHT4x (SHT40 / SHT41 / SHT45) on the default I2C pins.
// Driver: tinygo.org/x/drivers/sht4x (address 0x44).
//
// Wiring:
//
//	sensor SDA -- D4 (GPIO6)
//	sensor SCL -- D5 (GPIO7)
//	sensor VDD -- 3V3
//	sensor GND -- GND
//
// The breakout boards usually include the I2C pull-ups. A bare sensor wants
// external 4.7 kΩ pull-ups to 3V3 on SDA and SCL.
package main

import (
	"time"

	"machine"

	"tinygo.org/x/drivers/sht4x"
)

func main() {
	time.Sleep(time.Second)
	machine.I2C0.Configure(machine.I2CConfig{Frequency: 400_000})
	sensor := sht4x.New(machine.I2C0)

	println("SHT4x on I2C0 SDA=D4 SCL=D5 addr=0x44")
	for {
		temp, hum, err := sensor.ReadTemperatureHumidity()
		if err != nil {
			println("read error:", err.Error())
		} else {
			println("temperature", formatMilli(temp), "C")
			println("humidity", formatMilli(hum), "%RH")
		}
		time.Sleep(2 * time.Second)
	}
}

// formatMilli prints a value that is scaled by 1000, with three decimal places.
func formatMilli(v int32) string {
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	whole := v / 1000
	frac := v % 1000
	return sign + itoa(whole) + "." + pad3(frac)
}

func pad3(v int32) string {
	s := itoa(v)
	for len(s) < 3 {
		s = "0" + s
	}
	return s
}

func itoa(v int32) string {
	if v == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
