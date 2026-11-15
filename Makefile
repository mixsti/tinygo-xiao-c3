# USB serial CLI for the Seeed Studio XIAO ESP32-C3.
# Pinned toolchain: TinyGo 0.44.0.

TINYGO ?= tinygo
TARGET ?= xiao-esp32c3
PORT ?=

PORT_ARG := $(if $(PORT),-port=$(PORT),)

.PHONY: all build test flash monitor flash-monitor

all: build

build: build/xiao.bin

build/xiao.bin: main.go line.go wifi.go store.go reboot.go console/parse.go console/record.go go.mod go.sum
	@mkdir -p build
	$(TINYGO) build -target=$(TARGET) -o $@ .

test:
	go test ./console

flash:
	$(TINYGO) flash -target=$(TARGET) $(PORT_ARG) .

monitor:
	$(TINYGO) monitor -target=$(TARGET) $(PORT_ARG)

flash-monitor:
	$(TINYGO) flash -target=$(TARGET) -monitor $(PORT_ARG) .
