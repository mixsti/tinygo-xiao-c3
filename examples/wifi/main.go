//go:build tinygo

// Connect to Wi-Fi and perform one HTTP GET.
//
// Credentials are injected at link time so they are not stored in the repo:
//
//	tinygo flash -target=xiao-esp32c3 \
//	  -ldflags="-X main.ssid=myssid -X main.password=mypass" \
//	  -monitor ./examples/wifi
//
// Or: copy wifi.secrets.example to wifi.secrets and run `make flash EXAMPLE=wifi`.
//
// This uses net/http, which allocates per request. Fine for a one-shot GET.
// For a server that stays up, prefer the httphi examples in tinygo.org/x/espradio.
package main

import (
	"io"
	"net/http"
	"time"

	"tinygo.org/x/drivers/netdev"
	nl "tinygo.org/x/drivers/netlink"
	link "tinygo.org/x/espradio/netlink"
)

var (
	ssid     string
	password string
	rawURL   = "http://httpbin.org/get"
)

func main() {
	time.Sleep(2 * time.Second)
	if ssid == "" {
		println("ssid is empty.")
		println("flash with -ldflags=\"-X main.ssid=... -X main.password=...\"")
		println("or set WIFI_SSID and WIFI_PASS (see wifi.secrets.example)")
		for {
			time.Sleep(time.Second)
		}
	}

	dev := link.Esplink{}
	netdev.UseNetdev(&dev)

	println("connecting to", ssid)
	err := dev.NetConnect(&nl.ConnectParams{
		Ssid:       ssid,
		Passphrase: password,
	})
	if err != nil {
		println("wifi connect:", err.Error())
		return
	}
	addr, err := dev.Addr()
	if err != nil {
		println("address:", err.Error())
	} else {
		println("ip", addr.String())
	}

	println("GET", rawURL)
	resp, err := http.Get(rawURL)
	if err != nil {
		println("http:", err.Error())
		return
	}
	println("status", resp.Status)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		println("body:", err.Error())
		return
	}
	if len(body) > 240 {
		body = body[:240]
	}
	println(string(body))
	println("done")
	for {
		time.Sleep(time.Hour)
	}
}
