//go:build tinygo

// Scan for BLE advertisements and print address, RSSI, and local name.
//
// Same radio support as examples/ble-advertise. Scanning is the central role;
// this program does not advertise and does not bring up Wi-Fi.
package main

import (
	"time"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	time.Sleep(time.Second)
	println("enabling BLE")
	must("enable BLE", adapter.Enable())

	println("scanning")
	err := adapter.Scan(func(_ *bluetooth.Adapter, device bluetooth.ScanResult) {
		println("device", device.Address.String(), "rssi", device.RSSI, "name", device.LocalName())
	})
	must("scan", err)
}

func must(action string, err error) {
	if err != nil {
		for {
			println(action+":", err.Error())
			time.Sleep(time.Second)
		}
	}
}
