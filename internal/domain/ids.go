package domain

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode"
)

// NewID builds a stable, human-friendly id from a display name: a slug derived
// from the name plus a short random suffix to keep it unique.
func NewID(name string) string {
	s := slugify(name)
	if s == "" {
		s = "item"
	}
	return s + "-" + randSuffix(4)
}

// slugify lowercases name and replaces runs of non-alphanumeric characters with
// single hyphens, trimming leading/trailing hyphens.
func slugify(name string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen && b.Len() > 0 {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// randSuffix returns n hex characters of randomness.
func randSuffix(n int) string {
	buf := make([]byte, (n+1)/2)
	if _, err := rand.Read(buf); err != nil {
		return "0000"[:n]
	}
	return hex.EncodeToString(buf)[:n]
}
