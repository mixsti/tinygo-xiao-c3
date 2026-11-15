//go:build tinygo

// Interactive CLI on the XIAO ESP32-C3 USB-Serial-JTAG console.
//
//	tinygo flash -target=xiao-esp32c3 -monitor .
//
// `tinygo monitor` delivers Enter as CR. Backspace and DEL delete the last
// character. Wi-Fi credentials are never printed back in full.
package main

import (
	"strconv"
	"strings"
	"time"

	"machine"

	"github.com/mixsti/tinygo-xiao-c3/console"
)

var flashSaved bool

func main() {
	time.Sleep(time.Second)
	println("XIAO ESP32-C3 console")
	println("USB-Serial-JTAG. Type help.")
	loadSaved()
	if flashSaved && ssid != "" {
		println("saved wifi config, connecting")
		cmdConnect()
	}
	for {
		prompt()
		line := strings.TrimSpace(readLine())
		if line == "" {
			continue
		}
		dispatch(console.Split(line))
	}
}

func dispatch(args []string) {
	if len(args) == 0 {
		return
	}
	switch args[0] {
	case "help", "?":
		printHelp()
	case "wifi":
		wifiCmd(args[1:])
	case "ping":
		if len(args) != 2 {
			println("usage: ping <host>")
			return
		}
		cmdPing(args[1])
	case "dns":
		if len(args) != 2 {
			println("usage: dns <host>")
			return
		}
		cmdDNS(args[1])
	case "tcp":
		tcpCmd(args[1:])
	case "config":
		configCmd(args[1:])
	case "reboot":
		reboot()
	default:
		println("unknown command. type help")
	}
}

func wifiCmd(args []string) {
	if len(args) == 0 {
		println("usage: wifi scan|set|connect|disconnect|status")
		return
	}
	switch args[0] {
	case "scan":
		cmdScan()
	case "set":
		if len(args) < 2 {
			println("usage: wifi set <ssid> <password>")
			return
		}
		cmdSet(args[1], strings.Join(args[2:], " "))
	case "connect":
		cmdConnect()
	case "disconnect":
		cmdDisconnect()
	case "status":
		cmdStatus()
	default:
		println("unknown wifi command. type help")
	}
}

func tcpCmd(args []string) {
	if len(args) < 1 || len(args) > 2 {
		println("usage: tcp <host> [port]")
		return
	}
	port := 80
	if len(args) == 2 {
		n, err := strconv.Atoi(args[1])
		if err != nil || n < 1 || n > 65535 {
			println("port must be 1..65535")
			return
		}
		port = n
	}
	cmdTCP(args[0], uint16(port))
}

func configCmd(args []string) {
	if len(args) != 1 {
		println("usage: config show|save|clear")
		return
	}
	switch args[0] {
	case "show":
		if ssid == "" {
			println("ssid (none)")
		} else {
			println("ssid", ssid)
		}
		println("password", maskPassword(password))
		where := "ram"
		if flashSaved {
			where = "ram + flash"
		}
		println("stored", where)
	case "save":
		saveConfig()
	case "clear":
		clearConfig()
	default:
		println("usage: config show|save|clear")
	}
}

func printHelp() {
	println("help                         this list")
	println("wifi scan                    SSID and RSSI")
	println("wifi set <ssid> <password>   keep credentials in RAM")
	println("wifi connect                 join the AP, 3 attempts")
	println("wifi disconnect")
	println("wifi status                  state, IP, gateway, RSSI, MAC")
	println("ping <host>                  ICMP echo")
	println("dns <host>                   resolve an A record")
	println("tcp <host> [port]            TCP connect test, default port 80")
	println("config show                  SSID and masked password")
	println("config save                  write credentials to flash")
	println("config clear                 erase flash and RAM credentials")
	println("reboot")
	println("scan results have no channel or auth: espradio v0.5.0 does not export them")
}

func maskPassword(p string) string {
	return console.MaskSecret(p)
}

func reboot() {
	println("rebooting")
	time.Sleep(50 * time.Millisecond)
	machine.FlushSerial()
	// RTC_CNTL_OPTIONS0 SW_SYS_RST, the same bit ESP-IDF esp_restart sets.
	espReset()
}
