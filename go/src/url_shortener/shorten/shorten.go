package shorten

import (
	"crypto/sha1"
	"fmt"
)

// Shorten function returns a unique shorten form for a URL
func Shorten(longURL string) string {
	hasher := sha1.New()

	hasher.Write([]byte(longURL))
	sum := hasher.Sum(nil)

	shortURL := fmt.Sprintf("%x", sum)[:7]

	return shortURL
}
