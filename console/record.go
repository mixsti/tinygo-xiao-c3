package console

import "errors"

// On-flash credential record. Fixed size so it can be written as one
// word-aligned blob into a single 4 KiB erase block.
//
//	0:4    magic "XC31"
//	4      ssid length
//	5      password length
//	6:38   ssid bytes (32)
//	38:102 password bytes (64)
//	102:104 checksum, little endian, over bytes [0:102]

const (
	Magic      = "XC31"
	MaxSSID    = 32
	MaxPass    = 64
	RecordSize = 104
)

var (
	ErrSSIDTooLong = errors.New("ssid longer than 32 bytes")
	ErrPassTooLong = errors.New("password longer than 64 bytes")
	ErrEmptySSID   = errors.New("ssid is empty")
)

// Credentials is the Wi-Fi station identity stored in RAM and, when saved, in flash.
type Credentials struct {
	SSID     string
	Password string
}

// Encode serializes c. The password is stored so the board can reconnect;
// callers must not print the encoded bytes.
func Encode(c Credentials) ([]byte, error) {
	if c.SSID == "" {
		return nil, ErrEmptySSID
	}
	if len(c.SSID) > MaxSSID {
		return nil, ErrSSIDTooLong
	}
	if len(c.Password) > MaxPass {
		return nil, ErrPassTooLong
	}
	buf := make([]byte, RecordSize)
	copy(buf[0:4], Magic)
	buf[4] = byte(len(c.SSID))
	buf[5] = byte(len(c.Password))
	copy(buf[6:38], c.SSID)
	copy(buf[38:102], c.Password)
	sum := checksum(buf[:102])
	buf[102] = byte(sum)
	buf[103] = byte(sum >> 8)
	return buf, nil
}

// Decode validates magic and checksum. ok is false for a blank or corrupt block.
func Decode(b []byte) (Credentials, bool) {
	if len(b) < RecordSize {
		return Credentials{}, false
	}
	if string(b[0:4]) != Magic {
		return Credentials{}, false
	}
	sum := uint16(b[102]) | uint16(b[103])<<8
	if checksum(b[:102]) != sum {
		return Credentials{}, false
	}
	ssidLen := int(b[4])
	passLen := int(b[5])
	if ssidLen == 0 || ssidLen > MaxSSID || passLen > MaxPass {
		return Credentials{}, false
	}
	return Credentials{
		SSID:     string(b[6 : 6+ssidLen]),
		Password: string(b[38 : 38+passLen]),
	}, true
}

func checksum(b []byte) uint16 {
	var s uint16
	for _, c := range b {
		s = s*33 + uint16(c)
	}
	return s
}
