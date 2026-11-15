# TinyGo CLI for the Seeed XIAO ESP32-C3

One program: a command line on the native USB-Serial-JTAG port. It configures Wi-Fi, connects, and can store the credentials in on-chip flash. Nothing in git contains an SSID or password.

Checked against **TinyGo 0.44.0** (Go 1.25), `tinygo.org/x/espradio` v0.5.0, and `github.com/soypat/lneto` v0.4.0. CI compiles for `xiao-esp32c3` and does not flash a board.

## Prerequisites

- [Go 1.25](https://go.dev/dl/)
- [TinyGo 0.44.0](https://github.com/tinygo-org/tinygo/releases/tag/v0.44.0)
- A data USB-C cable. Charge-only cables never enumerate.

`tinygo flash` programs the chip over USB-Serial-JTAG (VID:PID `303a:1001`). A separate esptool install is not required. On Linux, your user needs access to `/dev/ttyACM0` (the `dialout` group).

```sh
tinygo version   # tinygo version 0.44.0, using go version go1.25.x
```

## Flash and open the console

```sh
make flash-monitor
# or
tinygo flash -target=xiao-esp32c3 -monitor .
```

`make monitor` is `tinygo monitor -target=xiao-esp32c3`. With one board plugged in, TinyGo finds the port. Otherwise:

```sh
make monitor PORT=/dev/ttyACM0          # Linux
make monitor PORT=/dev/cu.usbmodem1101  # macOS
```

Baud is 115200. The USB-Serial-JTAG hardware ignores it. Enter arrives as CR, which the CLI accepts. Ctrl-C quits the monitor. It does not reset the board.

The console is the USB port, not UART0 on D6/D7.

## Commands

```
xiao> help
wifi scan
wifi set myssid my-password
wifi connect
wifi status
ping 1.1.1.1
dns example.com
tcp example.com 80
config show
config save
config clear
reboot
```

`wifi set` keeps the SSID and password in RAM. `wifi connect` uses that copy and tries the association three times. `config save` writes them to flash. The next boot loads that record and connects on its own. `config clear` erases both.

`config show` prints `******** (N chars)`, never the stored password. Keystrokes are echoed as you type, including `wifi set`.

An SSID with spaces needs quotes: `wifi set "my network" secret`. The password is the rest of the line.

`wifi scan` prints SSID and RSSI. espradio v0.5.0 does not export channel or auth mode, and the CLI says so. A scan can drop the association. Run `wifi connect` again if status looks wrong.

`wifi status` prints state, SSID, IPv4, gateway, RSSI, and the station MAC. `ping` is ICMP echo. If a host drops ICMP, `tcp <host> 80` or `tcp <host>:<port>` checks TCP. A LAN target is resolved with ARP before the handshake, and the SYN is retried.

## Saved credentials

`config save` erases the last 4 KiB of the 4 MB flash and writes one record (magic `XC31`, SSID, password, checksum) via `machine.Flash`. TinyGo 0.44 maps that device to the gap after the program. A normal reflash does not need that last sector. `esptool erase_flash` deletes the record.

The password is stored in the clear. Anyone who can read the chip can read it.

## Pinout

From the [TinyGo XIAO ESP32-C3 page](https://tinygo.org/docs/reference/microcontrollers/featured/xiao-esp32c3/) for 0.44.0. This program does not drive these pins. The CH LED is the charger indicator, not a GPIO.

| Pin | GPIO | Also |
| --- | --- | --- |
| D0 | GPIO2 | A0 |
| D1 | GPIO3 | A1 |
| D2 | GPIO4 | A2 |
| D3 | GPIO5 | A3, avoid while Wi-Fi is up |
| D4 | GPIO6 | SDA |
| D5 | GPIO7 | SCL |
| D6 | GPIO21 | UART TX, not the USB console |
| D7 | GPIO20 | UART RX, not the USB console |
| D8 | GPIO8 | SPI SCK |
| D9 | GPIO9 | SPI SDI, **BOOT strap** (low at reset enters the ROM) |
| D10 | GPIO10 | SPI SDO |

## Troubleshooting

**No serial port.** Use a data cable. `lsusb` should show `303a:1001`.

**Flash fails.** Hold BOOT, tap RESET, release BOOT, then flash again. BOOT is D9 / GPIO9. Anything holding D9 low at reset keeps the chip in the download ROM.

**Monitor is silent.** Wait a second after flashing, then `make monitor`. Press RESET. The CLI waits about a second so USB can reattach, then prints a banner.

**`reboot` hangs.** It sets `RTC_CNTL` `SW_SYS_RST`. If the chip stays there, press RESET.

## Layout

```
main.go line.go wifi.go store.go reboot.go   the CLI
console/                                     record format and line splitting, host-tested
```

`//go:build tinygo` keeps `go test` from compiling `machine` on the host. `make test` runs the console tests. `make build` is the firmware build.
