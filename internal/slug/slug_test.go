package slug

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	g := New()

	got, err := g.Generate()
	if err != nil {
		t.Fatalf("Generate(): %v", err)
	}

	if len(got) != DefaultLength {
		t.Errorf("got %q (%d chars), want %d chars", got, len(got), DefaultLength)
	}
}
