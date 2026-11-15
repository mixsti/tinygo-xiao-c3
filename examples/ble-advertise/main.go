//go:build tinygo

// Advertise as a BLE peripheral named XIAO-C3.
//
// Requires TinyGo 0.42 or later. Verified to compile with TinyGo 0.44.0 and
// tinygo.org/x/bluetooth v0.17.0. The target sets the espradio build tag, so
// the stack uses the on-chip controller.
//
// BLE-only: this program does not start Wi-Fi. If you combine both radios in
// one program, call espradio.Enable before adapter.Enable. Wi-Fi must come first.
package main

import (
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	time.Sleep(time.Second)
	must("enable BLE", adapter.Enable())

	adv := adapter.DefaultAdvertisement()
	must("configure advertisement", adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: "XIAO-C3",
	}))
	must("start advertisement", adv.Start())

	address, err := adapter.Address()
	if err != nil {
		println("address:", err.Error())
	}
	println("advertising as XIAO-C3")
	for {
		if err == nil {
			println("advertising", address.MAC.String())
		} else {
			println("advertising")
		}
		time.Sleep(time.Second)
	}
}

func must(action string, err error) {
	if err != nil {
		for {
			println(action+":", err.Error())
			time.Sleep(time.Second)
		}
	}
}
