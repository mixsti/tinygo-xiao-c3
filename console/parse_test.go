package console

import "testing"

func TestSplit(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"wifi status", []string{"wifi", "status"}},
		{`wifi set "my net" secret`, []string{"wifi", "set", "my net", "secret"}},
		{"wifi set cafe p a s", []string{"wifi", "set", "cafe", "p", "a", "s"}},
		{`ping "one two"`, []string{"ping", "one two"}},
		{"", nil},
	}
	for _, tc := range cases {
		got := Split(tc.in)
		if len(got) != len(tc.want) {
			t.Fatalf("Split(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("Split(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		}
	}
}

func TestMaskSecretOmitsPassword(t *testing.T) {
	const secret = "hunter2-hunter2"
	got := MaskSecret(secret)
	if got == secret || contains(got, secret) {
		t.Fatalf("mask leaked secret: %q", got)
	}
	if MaskSecret("") != "(none)" {
		t.Fatalf("empty: %q", MaskSecret(""))
	}
	if got != "******** (15 chars)" {
		t.Fatalf("mask = %q", got)
	}
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (s == sub || len(s) > 0 && (indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
