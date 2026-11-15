# TinyGo starter for the Seeed Studio XIAO ESP32-C3.
# Pinned toolchain: TinyGo 0.44.0 (see README). Override TINYGO if it is not on PATH.

TINYGO ?= tinygo
TARGET ?= xiao-esp32c3
EXAMPLE ?= blink
PORT ?=

# Optional local credentials. wifi.secrets is gitignored.
-include wifi.secrets
WIFI_SSID ?=
WIFI_PASS ?=

# One shell word so both -X assignments reach the linker.
LDFLAGS_WIFI := -ldflags="-X main.ssid=$(WIFI_SSID) -X main.password=$(WIFI_PASS)"

EXAMPLES := blink serial button sht4x wifi ble-advertise ble-scan cli

PORT_ARG := $(if $(PORT),-port=$(PORT),)
WIFI_LDFLAGS_FOR = $(if $(filter wifi,$(1)),$(LDFLAGS_WIFI),)

.PHONY: all build test flash monitor flash-monitor help list

all: build

help:
	@echo "Targets:"
	@echo "  make build                  build every example for $(TARGET)"
	@echo "  make test                   host tests (parser, flash record)"
	@echo "  make flash EXAMPLE=blink    flash one example"
	@echo "  make monitor                USB-Serial-JTAG console (115200)"
	@echo "  make flash-monitor          flash EXAMPLE, then open the monitor"
	@echo "  make list"
	@echo "Variables: EXAMPLE ($(EXAMPLE)), PORT, WIFI_SSID, WIFI_PASS, TINYGO, TARGET"

list:
	@printf '%s\n' $(EXAMPLES)

build: $(addprefix build/,$(addsuffix .bin,$(EXAMPLES)))

build/wifi.bin: examples/wifi/main.go go.mod go.sum
	@mkdir -p build
	$(TINYGO) build -target=$(TARGET) -o $@ $(LDFLAGS_WIFI) ./examples/wifi

build/cli.bin: examples/cli/main.go examples/cli/line.go examples/cli/wifi.go examples/cli/store.go examples/cli/reboot.go console/parse.go console/record.go

build/%.bin: examples/%/main.go go.mod go.sum
	@mkdir -p build
	$(TINYGO) build -target=$(TARGET) -o $@ ./examples/$*

test:
	go test ./...

flash:
	$(TINYGO) flash -target=$(TARGET) $(PORT_ARG) $(call WIFI_LDFLAGS_FOR,$(EXAMPLE)) ./examples/$(EXAMPLE)

monitor:
	$(TINYGO) monitor -target=$(TARGET) $(PORT_ARG)

flash-monitor:
	$(TINYGO) flash -target=$(TARGET) -monitor $(PORT_ARG) $(call WIFI_LDFLAGS_FOR,$(EXAMPLE)) ./examples/$(EXAMPLE)
