//go:build tinygo

package main

import (
	"context"
	"errors"
	"net/netip"
	"strconv"
	"time"

	"github.com/soypat/lneto"
	"github.com/soypat/lneto/dns"
	"github.com/soypat/lneto/tcp"
	"github.com/soypat/lneto/x/xnet"
	"tinygo.org/x/espradio"
)

// Radio state for the CLI. The stack is the same shape as espradio's
// connect-and-dhcp example, which is what exposes the gateway address and
// ICMP echo. espradio/netlink.Esplink does not return the gateway.

var (
	radioReady bool
	stackReady bool
	linkUp     bool
	haveLease  bool
	ssid       string
	password   string
	joinedSSID string
	ip         netip.Addr
	gateway    netip.Addr
	stack      xnet.Netstack
	netdev     *espradio.NetDev
	localPort  uint16 = 41000
)

func backoff(uint) time.Duration { return 5 * time.Millisecond }

func ensureRadio() error {
	err := espradio.Enable(espradio.Config{Logging: espradio.LogLevelError})
	if err != nil && !errors.Is(err, espradio.ErrAlreadyEnabled) {
		return err
	}
	if radioReady {
		return nil
	}
	if err := espradio.Start(); err != nil {
		return err
	}
	radioReady = true
	return nil
}

func cmdScan() {
	if err := ensureRadio(); err != nil {
		println("radio:", err.Error())
		return
	}
	println("scanning (this can take several seconds)")
	aps, err := espradio.Scan()
	if err != nil {
		println("scan:", err.Error())
		return
	}
	if len(aps) == 0 {
		println("no access points")
	}
	for i, ap := range aps {
		println(strconv.Itoa(i+1), ap.SSID, "rssi", strconv.Itoa(ap.RSSI))
	}
	// espradio v0.5.0 AccessPoint is only SSID and RSSI. The underlying
	// wifi_ap_record_t also has channel and authmode, but they are not exported.
	println("note: channel and auth mode are not in espradio v0.5.0 scan results")
	if linkUp {
		println("note: a scan can drop the association; run wifi connect again if needed")
	}
}

func cmdSet(name, pass string) {
	if name == "" {
		println("usage: wifi set <ssid> <password>")
		return
	}
	if len(name) > 32 || len(pass) > 64 {
		println("ssid max 32 bytes, password max 64")
		return
	}
	ssid = name
	password = pass
	println("ssid", ssid)
	println("password", maskPassword(password))
	println("stored in RAM. config save writes flash. wifi connect uses RAM.")
}

func cmdConnect() {
	if ssid == "" {
		println("no ssid. wifi set <ssid> <password> first")
		return
	}
	if err := ensureRadio(); err != nil {
		println("radio:", err.Error())
		return
	}
	println("connecting to", ssid)
	err := espradio.Connect(espradio.STAConfig{SSID: ssid, Password: password})
	if err != nil {
		linkUp = false
		println("connect:", err.Error())
		return
	}
	linkUp = true
	joinedSSID = ssid
	println("associated")
	if err := ensureStack(); err != nil {
		println("stack:", err.Error())
		return
	}
	// Drop any previous lease so a second `wifi connect` can request again.
	stack.EnableDHCP(context.Background(), false, netip.Addr{})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	addr, gw, _, err := stack.EnableDHCP(ctx, true, netip.Addr{})
	cancel()
	if err != nil {
		haveLease = false
		println("dhcp:", err.Error())
		return
	}
	ip = addr
	gateway = gw
	haveLease = true
	println("ip", ip.String())
	println("gateway", gateway.String())
}

func cmdDisconnect() {
	if !radioReady {
		println("radio is down")
		return
	}
	if err := espradio.Disconnect(); err != nil {
		println("disconnect:", err.Error())
		return
	}
	linkUp = false
	haveLease = false
	ip = netip.Addr{}
	gateway = netip.Addr{}
	println("disconnected")
}

func cmdStatus() {
	state := "down"
	switch {
	case linkUp && haveLease:
		state = "connected"
	case linkUp:
		state = "associated"
	case radioReady:
		state = "radio up"
	}
	println("state", state)
	if joinedSSID != "" && linkUp {
		println("ssid", joinedSSID)
	} else if ssid != "" {
		println("ssid", ssid, "(not associated)")
	} else {
		println("ssid (none)")
	}
	if haveLease {
		println("ip", ip.String())
		println("gateway", gateway.String())
	} else {
		println("ip (none)")
		println("gateway (none)")
	}
	if radioReady {
		if rssi, err := espradio.RSSI(); err != nil {
			println("rssi (not connected)")
		} else {
			println("rssi", strconv.Itoa(rssi), "dBm")
		}
		if mac, err := espradio.StationMAC(); err != nil {
			println("mac", err.Error())
		} else {
			println("mac", formatMAC(mac))
		}
	} else {
		println("rssi (radio down)")
		println("mac (radio down)")
	}
}

func cmdDNS(host string) {
	if !haveLease {
		println("not connected")
		return
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		println(addr.String())
		return
	}
	addr, err := lookup(host)
	if err != nil {
		println("dns:", err.Error())
		return
	}
	println(host, addr.String())
}

func cmdPing(host string) {
	addr, err := resolve(host)
	if err != nil {
		println("ping:", err.Error())
		return
	}
	block := stack.StackAsync().StackBlocking(backoff)
	start := time.Now()
	rtt, err := block.DoPing(addr, 3*time.Second)
	if err != nil {
		println("ping", host, err.Error(), "after", time.Since(start).String())
		println("icmp echo may be filtered; try: tcp", host, "80")
		return
	}
	println("ping", addr.String(), "time", rtt.String())
}

func cmdTCP(host string, port uint16) {
	addr, err := resolve(host)
	if err != nil {
		println("tcp:", err.Error())
		return
	}
	var conn tcp.Conn
	if err := conn.Configure(tcp.ConnConfig{
		RxBuf:             make([]byte, 256),
		TxBuf:             make([]byte, 256),
		TxPacketQueueSize: 4,
	}); err != nil {
		println("tcp:", err.Error())
		return
	}
	localPort++
	if localPort < 41000 {
		localPort = 41000
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	block := stack.StackAsync().StackBlocking(backoff)
	start := time.Now()
	err = block.DoDialTCP(ctx, &conn, localPort, netip.AddrPortFrom(addr, port), 5*time.Second)
	if err != nil {
		println("tcp", host+":"+strconv.Itoa(int(port)), err.Error())
		conn.Abort()
		return
	}
	println("tcp connected", addr.String(), "port", strconv.Itoa(int(port)), "in", time.Since(start).String())
	conn.Close()
}

func resolve(host string) (netip.Addr, error) {
	if !haveLease {
		return netip.Addr{}, errNotConnected
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		return addr, nil
	}
	return lookup(host)
}

func lookup(host string) (netip.Addr, error) {
	var name dns.Name
	if err := name.Parse(host); err != nil {
		return netip.Addr{}, err
	}
	var addrs [4]netip.Addr
	block := stack.StackAsync().StackBlocking(backoff)
	n, err := block.DoLookupIP(addrs[:], name, 5*time.Second)
	if err != nil {
		return netip.Addr{}, err
	}
	if n == 0 {
		return netip.Addr{}, errNoAddress
	}
	return addrs[0], nil
}

func ensureStack() error {
	if stackReady {
		return nil
	}
	if netdev == nil {
		nd, err := espradio.StartNetDev()
		if err != nil {
			return err
		}
		netdev = nd
	}
	hw, err := netdev.HardwareAddr6()
	if err != nil {
		return err
	}
	err = stack.Reset(xnet.NetstackConfig{
		StackConfig: xnet.StackConfig{
			Hostname:          "xiao-c3",
			RandSeed:          time.Now().UnixNano() | 1,
			HardwareAddress:   hw,
			MTU:               1500,
			MaxActiveUDPPorts: 4,
			MaxActiveTCPPorts: 2,
			ICMPQueueLimit:    4,
			PassivePeers:      8,
		},
		BackoffStack: backoff,
		TCPPoolConfig: xnet.TCPPoolConfig{
			PoolSize:   2,
			QueueSize:  2,
			TxBufSize:  1024,
			RxBufSize:  512,
			NewBackoff: func() lneto.BackoffStrategy { return backoff },
		},
	})
	if err != nil {
		return err
	}
	if err := stack.EnableICMP(true); err != nil {
		return err
	}
	go func() {
		err := espradio.RunStack(context.Background(), netdev, &stack)
		println("stack stopped:", err.Error())
		stackReady = false
	}()
	stackReady = true
	return nil
}

func formatMAC(m [6]byte) string {
	const hexdigits = "0123456789abcdef"
	var b [17]byte
	for i := 0; i < 6; i++ {
		b[i*3] = hexdigits[m[i]>>4]
		b[i*3+1] = hexdigits[m[i]&0x0f]
		if i < 5 {
			b[i*3+2] = ':'
		}
	}
	return string(b[:])
}

type cliError string

func (e cliError) Error() string { return string(e) }

const (
	errNotConnected cliError = "not connected"
	errNoAddress    cliError = "no address"
)
