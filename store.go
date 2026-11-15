//go:build tinygo

package main

import (
	"machine"

	"github.com/mixsti/tinygo-xiao-c3/console"
)

// Credentials live in RAM until `config save`. The saved copy is one record
// in the last 4 KiB erase block of machine.Flash.
//
// On ESP32-C3, TinyGo 0.44 maps machine.Flash to the gap between the end of
// the program and the end of the 4 MB flash (linker symbols __flash_data_start
// and __flash_data_end). The last block of that region is the last 4 KiB of
// the chip, so a normal image rewrite does not have to touch it. A full-chip
// erase does. There is no NVS partition API; this is a single raw record.

func loadSaved() {
	off, ok := configBlock()
	if !ok {
		println("flash: no writable region, credentials stay in RAM")
		return
	}
	buf := make([]byte, console.RecordSize)
	if _, err := machine.Flash.ReadAt(buf, off); err != nil {
		println("flash read:", err.Error())
		return
	}
	cred, ok := console.Decode(buf)
	if !ok {
		return
	}
	ssid = cred.SSID
	password = cred.Password
	flashSaved = true
	println("loaded saved wifi ssid", ssid)
}

func saveConfig() {
	cred := console.Credentials{SSID: ssid, Password: password}
	buf, err := console.Encode(cred)
	if err != nil {
		println("config:", err.Error())
		return
	}
	off, ok := configBlock()
	if !ok {
		println("flash unavailable; credentials stay in RAM")
		return
	}
	erase := machine.Flash.EraseBlockSize()
	if err := machine.Flash.EraseBlocks(off/erase, 1); err != nil {
		println("flash erase:", err.Error())
		return
	}
	if _, err := machine.Flash.WriteAt(buf, off); err != nil {
		println("flash write:", err.Error())
		return
	}
	flashSaved = true
	println("saved to flash")
}

func clearConfig() {
	ssid = ""
	password = ""
	flashSaved = false
	off, ok := configBlock()
	if !ok {
		println("cleared RAM; flash unavailable")
		return
	}
	erase := machine.Flash.EraseBlockSize()
	if err := machine.Flash.EraseBlocks(off/erase, 1); err != nil {
		println("flash erase:", err.Error())
		return
	}
	println("cleared RAM and flash")
}

func configBlock() (int64, bool) {
	size := machine.Flash.Size()
	erase := machine.Flash.EraseBlockSize()
	if erase <= 0 || size < erase {
		return 0, false
	}
	return size - erase, true
}
