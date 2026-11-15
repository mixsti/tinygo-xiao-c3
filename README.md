# TinyGo on the Seeed XIAO ESP32-C3

A small starter for the [Seeed Studio XIAO ESP32-C3](https://wiki.seeedstudio.com/XIAO_ESP32C3_Getting_Started/) (ESP32-C3, 4 MB flash, native USB-Serial-JTAG). Examples cover GPIO, the USB console, I2C, Wi-Fi, and BLE, plus an interactive serial CLI that can save Wi-Fi credentials on-chip.

This tree is for the **C3** target `xiao-esp32c3` only. The XIAO ESP32-C6, S3, and C5 are different TinyGo targets (or, for some of them, not supported yet) and are not wired up here.

## Toolchain this was checked against

| Piece | Version |
| --- | --- |
| TinyGo | **0.44.0** (28 Oct 2026, current stable) |
| Go | 1.25.x (TinyGo 0.44 is built for Go 1.25) |
| `tinygo.org/x/espradio` | v0.5.0 |
| `tinygo.org/x/bluetooth` | v0.17.0 |
| `tinygo.org/x/drivers` | v0.37.0 |
| `github.com/soypat/lneto` | v0.4.0 |

Every example below was compiled with `tinygo build -target=xiao-esp32c3` using that TinyGo. Nothing here was run on a board in CI: the workflow builds only.

What that compile actually exercises:

- Wi-Fi station connect, disconnect, scan, RSSI, and station MAC (`espradio` v0.5.0).
- DHCP address **and gateway**, DNS, ICMP echo, and a TCP connect, via the lneto stack that espradio runs (same shape as espradio's `connect-and-dhcp` example).
- BLE advertising and scanning (`tinygo.org/x/bluetooth` v0.17.0, `espradio` build tag, which the `xiao-esp32c3` target sets for you).
- On-chip flash as `machine.Flash` (present for `esp32c3` in TinyGo 0.44).
- The USB-Serial-JTAG console as `machine.Serial` (`serial: "usb"` on this target).

## Pinout

From the [TinyGo XIAO ESP32-C3 page](https://tinygo.org/docs/reference/microcontrollers/featured/xiao-esp32c3/) and `src/machine/board_xiao-esp32c3.go` in TinyGo 0.44.0.

| Silkscreen | GPIO | Also called | Notes |
| --- | --- | --- | --- |
| D0 | GPIO2 | A0 | Default external LED in `examples/blink` |
| D1 | GPIO3 | A1 | Default button in `examples/button` |
| D2 | GPIO4 | A2 | |
| D3 | GPIO5 | A3 | TinyGo marks GPIO5 "avoid when Wi-Fi is used" |
| D4 | GPIO6 | SDA | I2C data, `examples/sht4x` |
| D5 | GPIO7 | SCL | I2C clock |
| D6 | GPIO21 | UART TX | Hardware UART, **not** the USB console |
| D7 | GPIO20 | UART RX | Hardware UART, **not** the USB console |
| D8 | GPIO8 | SPI SCK | |
| D9 | GPIO9 | SPI SDI | **BOOT strap.** Low at reset enters the download ROM |
| D10 | GPIO10 | SPI SDO | |
| — | GPIO0 | ADC0 | Not on the header |
| — | GPIO1 | ADC1 | Not on the header |

3V3 and GND are on the other side of the header. Logic is 3.3 V.

There is **no user LED**. The CH LED is the LiPo charger indicator. Firmware cannot blink it.

## Wiring used by the examples

**External LED** (`examples/blink`), active high:

```
D0 (GPIO2) ---- 330 Ω to 1 kΩ ---- LED anode
LED cathode ---- GND
```

Use a series resistor. A direct LED-to-GPIO connection will overcurrent the pin. Change `ledPin` in `examples/blink/main.go` if you pick another pin. Do not use D9.

**Button** (`examples/button`), active low, internal pull-up:

```
D1 (GPIO3) ---- button ---- GND
```

**SHT40 / SHT41 / SHT45** (`examples/sht4x`), I2C address 0x44:

```
SDA ---- D4 (GPIO6)
SCL ---- D5 (GPIO7)
VDD ---- 3V3
GND ---- GND
```

Breakout boards usually include the pull-ups. A bare sensor needs about 4.7 kΩ from SDA and SCL to 3V3.

## Prerequisites

1. [Go 1.25](https://go.dev/dl/) or any later 1.25 patch.
2. [TinyGo 0.44.0](https://github.com/tinygo-org/tinygo/releases/tag/v0.44.0). Other 0.44 patch releases should be source-compatible; this repo was built with 0.44.0 exactly.

   ```sh
   # Linux x86_64, matching what CI installs
   curl -fsSL -o /tmp/tinygo.tar.gz \
     https://github.com/tinygo-org/tinygo/releases/download/v0.44.0/tinygo0.44.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf /tmp/tinygo.tar.gz
   export PATH="/usr/local/tinygo/bin:$PATH"
   tinygo version   # tinygo version 0.44.0
   ```

   Homebrew and the Windows zip are on the same release page. TinyGo must be the 0.44 build that uses Go 1.25 (`tinygo version` prints the Go version it found).

3. A data USB-C cable. Charge-only cables never enumerate.

You do **not** need a separate esptool or OpenOCD install to flash. `tinygo flash -target=xiao-esp32c3` talks to the on-chip USB-Serial-JTAG ROM (the target's flash method is `esp32jtag`, USB VID:PID `303a:1001`). esptool is only an optional fallback, described under troubleshooting.

Put your user in the `dialout` group on Linux if opening `/dev/ttyACM0` is denied, then log in again.

## Build, flash, monitor

```sh
make build                  # every example, output in build/*.bin
make test                   # host tests for the CLI record format
make flash EXAMPLE=blink    # tinygo flash -target=xiao-esp32c3
make monitor                # USB serial console
make flash-monitor EXAMPLE=cli
```

`make monitor` is `tinygo monitor -target=xiao-esp32c3`. With a single board plugged in, TinyGo finds the `303a:1001` port. If you have several boards:

```sh
make monitor PORT=/dev/ttyACM0
# macOS:  PORT=/dev/cu.usbmodem1101
# Windows: PORT=COM5
```

The baud setting is 115200. USB-Serial-JTAG ignores baud, but that is what the monitor opens.

`tinygo monitor` keeps a few keys for itself: Ctrl-C quits the monitor (it does not reboot the chip), Ctrl-Z suspends it. Enter arrives at the board as CR (`\r`), which the examples accept.

One example at a time, without make:

```sh
tinygo flash -target=xiao-esp32c3 -monitor ./examples/blink
```

## Examples

| Directory | What it does | Flash |
| --- | --- | --- |
| `examples/blink` | Toggles an external LED on D0 | `make flash EXAMPLE=blink` |
| `examples/serial` | Prints hello, then echoes lines | `make flash EXAMPLE=serial` |
| `examples/button` | Debounced button on D1 | `make flash EXAMPLE=button` |
| `examples/sht4x` | SHT4x temperature and humidity | `make flash EXAMPLE=sht4x` |
| `examples/wifi` | Connect and HTTP GET | needs credentials, below |
| `examples/ble-advertise` | Advertises the name `XIAO-C3` | `make flash EXAMPLE=ble-advertise` |
| `examples/ble-scan` | Prints nearby advertisers | `make flash EXAMPLE=ble-scan` |
| `examples/cli` | Interactive console, the one to start from | `make flash EXAMPLE=cli` |

Example programs carry `//go:build tinygo`, so `go test ./...` does not try to compile `machine` with the host toolchain. Firmware builds go through TinyGo, which sets that tag.

### Wi-Fi credentials

Nothing in git contains an SSID or password.

`examples/wifi` reads them from linker variables:

```sh
tinygo flash -target=xiao-esp32c3 -monitor \
  -ldflags="-X main.ssid=myssid -X main.password=mypass" \
  ./examples/wifi
```

Or copy the example file and use make (the copy is gitignored):

```sh
cp wifi.secrets.example wifi.secrets
# edit WIFI_SSID and WIFI_PASS
make flash-monitor EXAMPLE=wifi
```

Passwords with spaces, quotes, or make metacharacters (`#`, `$`) are a bad fit for that file. Use the CLI for those.

`examples/wifi` performs one `net/http` GET of `http://httpbin.org/get` (override with `-X main.rawURL=...`). `net/http` allocates per request. That is fine for a single GET and a poor fit for a server that stays up. For a long-running server, use the `httphi` examples in [espradio](https://github.com/tinygo-org/espradio).

### CLI

Flash `examples/cli`, open the monitor, and you get a `xiao>` prompt. Backspace and Delete erase the last character. Enter submits the line.

```
help
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

`wifi set` keeps the SSID and password in RAM. `wifi connect` uses that RAM copy. `config save` writes them to flash. `config clear` erases both. On the next boot the CLI loads a valid flash record back into RAM.

`config show` and the confirmation after `wifi set` print a masked password (`******** (N chars)`), never the stored secret. The line editor does echo keystrokes as you type `wifi set`, because this is a local debug console.

An SSID with spaces needs quotes: `wifi set "my network" secret`. The password is the rest of the line, so it may contain spaces.

`wifi scan` prints SSID and RSSI. It then prints a note, because that is the whole public result.

## Where credentials are stored

TinyGo 0.44 implements `machine.Flash` on the ESP32-C3: a `BlockDevice` over the flash between the end of the program and the end of the 4 MB chip (`__flash_data_start` .. `__flash_data_end` in `targets/esp32c3.ld`). There is no NVS or filesystem in this tree.

`config save` erases the **last 4 KiB sector** of that region (the last 4 KiB of the chip) and writes one 104-byte record: magic `XC31`, lengths, SSID, password, checksum. The password is stored in the clear. Anyone who can read the flash can read it. That is the tradeoff of a raw sector on a hobby board.

A normal `tinygo flash` rewrites the application image from the start of flash and does not need that last sector. `esptool erase_flash`, or any full-chip erase, deletes the record. If `machine.Flash.Size` is ever smaller than one erase block, `config save` refuses and the credentials stay in RAM.

## Radios

Both radios are on the one 2.4 GHz radio inside the ESP32-C3.

- Wi-Fi in this repo is `tinygo.org/x/espradio` v0.5.0: station join, scan, DHCP, DNS, ICMP, TCP. Soft-AP and ESP-NOW exist in espradio and are not wrapped here.
- BLE is `tinygo.org/x/bluetooth` v0.17.0. TinyGo has had this path since 0.42. The two BLE examples are separate programs, so each one only brings up BLE.
- If you later use both in **one** program, enable Wi-Fi first, then BLE:

  ```go
  espradio.Enable(espradio.Config{})
  adapter.Enable()
  ```

  `adapter.Enable` alone is enough for a BLE-only program. Calling `espradio.Enable` a second time returns `ErrAlreadyEnabled`; a failed `Enable` cannot be retried without a reset (`ErrEnableFailed`).

### What the CLI can and cannot ask the radio

These commands are implemented and compiled for `xiao-esp32c3`:

| Command | Behavior |
| --- | --- |
| `wifi scan` | SSID and RSSI for each AP |
| `wifi set` / `connect` / `disconnect` | Station mode |
| `wifi status` | state, SSID, IPv4, gateway, RSSI, station MAC |
| `ping <host>` | ICMP echo via lneto `StackBlocking.DoPing` |
| `dns <host>` | IPv4 lookup |
| `tcp <host> [port]` | TCP handshake, default port 80 |
| `config save` | last 4 KiB flash sector |
| `reboot` | sets `RTC_CNTL` `SW_SYS_RST` |

Not available in the published v0.5.0 API, so they are not faked:

- **Channel and auth mode in scan results.** `espradio.AccessPoint` only has `SSID` and `RSSI`. The ESP-IDF `wifi_ap_record_t` has channel and auth mode, and the driver reads that struct, but it does not copy those fields out. `wifi scan` says so on the console.
- **Associated-AP channel in `wifi status`.** `Connect` does not return it, and the public link callback is only up/down.
- **Gateway through `espradio/netlink.Esplink`.** That type exposes `Addr()` and not the router. The CLI uses the lower-level stack (`EnableDHCP` returns the gateway), which is why `wifi status` can print it. `examples/wifi` uses `Esplink` and only prints the station address.

ICMP is often dropped. If `ping` fails, `tcp <host> 80` (or 443) is the connectivity check. A scan can knock the station off the AP; run `wifi connect` again if `wifi status` looks wrong afterwards.

## Known limitations

- No user LED, and no programmable USB device class. The USB-Serial-JTAG port is a fixed CDC console. TinyGo's board page lists USBDevice as unsupported. `machine.Serial` is that console, not UART0 on D6/D7.
- D9 (GPIO9) is the BOOT strap. A button or an LED held low on D9 at reset stops the application from booting.
- D3 (GPIO5) is best left unused while Wi-Fi is up; that warning is in TinyGo's ESP32-C3 pin list.
- `examples/wifi` links `net/http` and is large (about 940 KB of flash image). The CLI, which stays on lneto, is about 600 KB. The chip has 4 MB of flash. Prefer `httphi` if you grow a server.
- Adding `-tags=noslog` drops lneto's default slog handler (it otherwise pulls in `encoding/json`). espradio documents this. The builds here do not pass that tag.
- Flash credentials are plaintext. See above.
- Builds are verified. Behavior on a real XIAO is not, because CI has no board.

## Troubleshooting

**`tinygo flash` cannot find a port, or the port never shows up**

- Use a data cable, and the USB-C plug on the board.
- Linux: `lsusb` should show `303a:1001` (Espressif USB JTAG/serial debug unit) and a `/dev/ttyACM*` node. `dmesg` after plugging in is the fastest check.
- macOS: `/dev/cu.usbmodem*`. Windows: a new COM port.

**Flash starts and then fails, or the board never runs**

The ROM only accepts a download if GPIO9 is low at reset.

1. Hold the BOOT button.
2. Tap RESET (or plug in USB while still holding BOOT).
3. Release BOOT.
4. Run `tinygo flash` again.

D9 is that BOOT button. If you have wired D9 to something that pulls it down, the chip will keep entering the ROM.

**Monitor is silent**

- Give the board a second after flashing, then `make monitor`. `tinygo flash -monitor` sometimes opens the port before USB re-enumerates.
- The console is USB-Serial-JTAG, not an external USB-UART adapter on D6/D7.
- Press RESET. The CLI waits about a second so the host can reattach, then prints a banner.

**`reboot` from the CLI does not come back**

`reboot` sets the `SW_SYS_RST` bit in `RTC_CNTL.OPTIONS0` and then waits. If that bit does not reset the chip, press the RESET button.

**Optional esptool path**

`tinygo build -target=xiao-esp32c3 -o app.bin ./examples/blink` writes an ESP32 image (first byte `0xE9`). The ROM on the C3 loads that image from offset `0x0`:

```sh
esptool.py --chip esp32c3 --port /dev/ttyACM0 --baud 460800 \
  write_flash -z 0x0 app.bin
```

Prefer `tinygo flash`. You only need esptool if you are recovering a board outside TinyGo. `esptool.py erase_flash` wipes the CLI's saved Wi-Fi record along with everything else.

## Layout

```
console/            credential record and line splitting (host-tested)
examples/blink      external LED
examples/serial     USB console echo
examples/button     debounced input
examples/sht4x      I2C SHT4x
examples/wifi       station + one HTTP GET
examples/ble-advertise
examples/ble-scan
examples/cli        USB command line
```

## License

Example code in this repo is MIT. See [LICENSE](LICENSE).

The Wi-Fi and BLE stacks it imports keep their own licenses (espradio is Apache-2.0 and BSD-3-Clause; Go Bluetooth and TinyGo drivers are BSD-3-Clause).
