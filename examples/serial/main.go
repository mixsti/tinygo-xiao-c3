//go:build tinygo

// Hello, then echo, on the native USB-Serial-JTAG console.
//
// machine.Serial is USBCDC on this target (serial: usb in the TinyGo target).
// It is not UART0 on D6/D7. Open it with `tinygo monitor` or `make monitor`.
// The monitor sends carriage return for Enter; this example accepts CR or LF.
package main

import (
	"time"

	"machine"
)

func main() {
	time.Sleep(time.Second) // let the host reattach after flashing
	println("hello from XIAO ESP32-C3")
	println("USB-Serial-JTAG console. Type a line and press Enter.")

	var buf [80]byte
	n := 0
	for {
		if machine.Serial.Buffered() == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		b, err := machine.Serial.ReadByte()
		if err != nil {
			continue
		}
		if b == '\r' || b == '\n' {
			if n == 0 {
				continue
			}
			println("echo:", string(buf[:n]))
			n = 0
			continue
		}
		if b == 0x08 || b == 0x7f {
			if n > 0 {
				n--
			}
			continue
		}
		if b < 0x20 || n == len(buf) {
			continue
		}
		buf[n] = b
		n++
		machine.Serial.WriteByte(b)
	}
}
