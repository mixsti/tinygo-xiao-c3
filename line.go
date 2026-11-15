//go:build tinygo

package main

import (
	"time"

	"machine"
)

const lineMax = 160

// readLine reads one edited line from the USB-Serial-JTAG console.
// Enter is CR (what `tinygo monitor` sends) or LF. Backspace is 0x08 or DEL.
func readLine() string {
	var buf [lineMax]byte
	n := 0
	skipLF := false
	for {
		if machine.Serial.Buffered() == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		b, err := machine.Serial.ReadByte()
		if err != nil {
			continue
		}
		if skipLF {
			skipLF = false
			if b == '\n' {
				continue
			}
		}
		switch b {
		case '\r':
			skipLF = true
			fallthrough
		case '\n':
			machine.Serial.Write([]byte("\r\n"))
			return string(buf[:n])
		case 0x08, 0x7f:
			if n > 0 {
				n--
				machine.Serial.Write([]byte("\b \b"))
			}
		default:
			if b < 0x20 || n == len(buf) {
				continue
			}
			buf[n] = b
			n++
			// Write flushes the USB FIFO. WriteByte waits for newline.
			machine.Serial.Write([]byte{b})
		}
	}
}

func prompt() {
	machine.Serial.Write([]byte("xiao> "))
}

func println(args ...string) {
	for i, a := range args {
		if i > 0 {
			machine.Serial.WriteByte(' ')
		}
		machine.Serial.Write([]byte(a))
	}
	machine.Serial.Write([]byte("\r\n"))
}
