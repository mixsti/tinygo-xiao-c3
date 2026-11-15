// Package console holds host-testable pieces of the USB serial CLI:
// line splitting and the on-flash credential record. The board-specific
// loop lives in examples/cli.
package console

import "strings"

// Split breaks a command line into fields. Double quotes group a field
// that contains spaces. A backslash escapes the next character.
func Split(line string) []string {
	var fields []string
	var b strings.Builder
	inQuote := false
	escape := false
	flush := func() {
		fields = append(fields, b.String())
		b.Reset()
	}
	for i := 0; i < len(line); i++ {
		c := line[i]
		if escape {
			b.WriteByte(c)
			escape = false
			continue
		}
		if c == '\\' {
			escape = true
			continue
		}
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if !inQuote && (c == ' ' || c == '\t') {
			if b.Len() > 0 {
				flush()
			}
			continue
		}
		b.WriteByte(c)
	}
	if b.Len() > 0 || inQuote {
		flush()
	}
	return fields
}

// MaskSecret hides a password. The result is a fixed star string plus the
// length, so the stored secret is never printed back in full.
func MaskSecret(secret string) string {
	if secret == "" {
		return "(none)"
	}
	return "******** (" + itoa(len(secret)) + " chars)"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [8]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
