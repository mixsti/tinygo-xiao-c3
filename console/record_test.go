package console

import (
	"bytes"
	"testing"
)

func TestRecordRoundTrip(t *testing.T) {
	in := Credentials{SSID: "lab", Password: "s3cret!"}
	buf, err := Encode(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(buf) != RecordSize {
		t.Fatalf("len %d", len(buf))
	}
	// The raw record holds the password. That is the flash format, not a print path.
	if !bytes.Contains(buf, []byte(in.Password)) {
		t.Fatal("encoded record should contain the password bytes")
	}
	got, ok := Decode(buf)
	if !ok || got != in {
		t.Fatalf("decode %+v ok=%v", got, ok)
	}
}

func TestRecordRejectsBlankAndCorrupt(t *testing.T) {
	if _, err := Encode(Credentials{}); err == nil {
		t.Fatal("empty ssid should fail")
	}
	blank := make([]byte, RecordSize)
	if _, ok := Decode(blank); ok {
		t.Fatal("blank flash should not decode")
	}
	buf, err := Encode(Credentials{SSID: "a", Password: ""})
	if err != nil {
		t.Fatal(err)
	}
	buf[10] ^= 0xff
	if _, ok := Decode(buf); ok {
		t.Fatal("corrupt record should not decode")
	}
	open, ok := Decode(mustEncode(t, Credentials{SSID: "open", Password: ""}))
	if !ok || open.Password != "" || open.SSID != "open" {
		t.Fatalf("open network: %+v %v", open, ok)
	}
}

func mustEncode(t *testing.T, c Credentials) []byte {
	t.Helper()
	b, err := Encode(c)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
