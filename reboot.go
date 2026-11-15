//go:build tinygo

package main

import "device/esp"

func espReset() {
	esp.RTC_CNTL.SetOPTIONS0_SW_SYS_RST(1)
	for {
	}
}
