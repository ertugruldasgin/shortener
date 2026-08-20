package link

import (
	"errors"
	"testing"
)

func TestValidateTarget(t *testing.T) {
	valid := []string{
		"https://example.com",
		"http://example.com/path?q=1",
		"https://sub.example.com:8443/a/b",
	}

	for _, raw := range valid {
		if err := ValidateTarget(raw); err != nil {
			t.Errorf("ValidateTarget(%q): unexpected error %v", raw, err)
		}
	}

	invalid := []string{
		"",
		"   ",
		"merhaba",
		"ftp://example.com",
		"javascript:alert(1)",
		"https://",
	}

	for _, raw := range invalid {
		err := ValidateTarget(raw)
		if !errors.Is(err, ErrInvalidTarget) {
			t.Errorf("ValidateTarget(%q): got %v, want ErrInvalidTarget", raw, err)
		}
	}
}
