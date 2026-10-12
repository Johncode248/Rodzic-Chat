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

// paymentReferenceCharset excludes visually-confusable characters (0/O,
// 1/I/L) since this code is meant to be read off a screen and typed by hand
// into a banking app's transfer title.
const paymentReferenceCharset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

const digitCharset = "0123456789"

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

// GeneratePin returns an 8-digit numeric PIN (each digit drawn
// independently via crypto/rand, zero-padded, e.g. "00841273" is a valid
// output) suitable for an authenticated user to type in by hand to join an
// org/co-parent/chat invite -- unlike GenerateToken, this is meant to be
// short enough to read off a screen and key in manually, not embedded in a
// URL. Callers are responsible for hashing it before storage (see
// library/pinhash) and for retrying generation on the rare uniqueIndex
// collision.
func GeneratePin() string {
	return randomString(digitCharset, 8)
}

// GeneratePaymentReference returns a short, unambiguous code (e.g.
// "RC-7F3K9Q") for a director to use as their platform-subscription bank
// transfer's title, so the operator can match an incoming transfer to the
// right organization by hand. Generated once per organization; not meant to
// be secret (a 32-symbol alphabet at length 6 is plenty to avoid accidental
// collisions between organizations, not to resist guessing).
func GeneratePaymentReference() string {
	return "RC-" + randomString(paymentReferenceCharset, 6)
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
