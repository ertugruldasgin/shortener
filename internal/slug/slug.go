// Package slug generates random URL-safe identifiers for shortened links.
package slug

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// DefaultLength is the slug length used by New. 6 characters carry 36 bits of entropy (~68.7B)
const DefaultLength = 6

// Generator produces random URL-safe slugs. Construct one with New.
type Generator struct {
	length int
}

func New() *Generator {
	return &Generator{length: DefaultLength}
}

// Generate returns a random slug. Uniqueness is not guaranteed, callers resolve collisions against a unique constraint in storage.
func (g *Generator) Generate() (string, error) {
	// base64 packs 3 bytes into 4 chars, so we need round up (length*3 / 4) via (divisor - 1) -> (length*3 + (4-1)) / 4
	byteLen := (g.length*3 + 3) / 4

	token := make([]byte, byteLen)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}

	encodedSlug := base64.RawURLEncoding.EncodeToString(token)[:g.length]

	return encodedSlug, nil
}
