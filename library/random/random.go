package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const loginCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
const passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const tokenCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(charset string, length int) string {
	var sb strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		sb.WriteByte(charset[n.Int64()])
	}
	return sb.String()
}

// GenerateLoginFromName builds a login candidate out of a person's name plus
// a random suffix, e.g. "jan.kowalski" -> "jan.kowalski.4f3a". Callers are
// responsible for retrying with a fresh call if the candidate collides with
// an existing login.
func GenerateLoginFromName(name string) string {
	slug := slugify(name)
	if slug == "" {
		slug = "konto"
	}

	return fmt.Sprintf("%s.%s", slug, randomString(loginCharset, 4))
}

func GeneratePassword() string {
	return randomString(passwordCharset, 10)
}

// GenerateToken returns a 32-character random token suitable for a
// hard-to-guess URL path segment (e.g. a registration link) — a 62-symbol
// alphabet at length 32 has ~190 bits of entropy, far beyond what's
// brute-forceable before a link's own expiry.
func GenerateToken() string {
	return randomString(tokenCharset, 32)
}

// slugify keeps ASCII letters/digits and collapses whitespace into dots.
// Polish diacritics (ą, ę, ł, ...) are dropped rather than transliterated —
// good enough for a login suffix, not meant to be a display name.
func slugify(name string) string {
	var sb strings.Builder
	lastWasDot := true
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			sb.WriteRune(r)
			lastWasDot = false
		case r == ' ' || r == '.' || r == '-':
			if !lastWasDot {
				sb.WriteRune('.')
				lastWasDot = true
			}
		}
	}

	return strings.Trim(sb.String(), ".")
}
