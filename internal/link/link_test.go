package link

import (
	"errors"
	"strings"
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

func TestValidateSlug(t *testing.T) {
	valid := []string{"ab", "mylink", "my-link_2", "A1"}

	for _, s := range valid {
		if err := ValidateSlug(s); err != nil {
			t.Errorf("ValidateSlug(%q): unexpected error %v", s, err)
		}
	}

	invalid := []string{
		"",                                // too short
		"a",                               // too short
		strings.Repeat("a", maxSlugLen+1), // too long
		"api",                             // reserved
		"API",                             // reserved, different case
		"my link",                         // space
		"my.link",                         // dot
		"türkçe",                          // non-ascii
	}

	for _, s := range invalid {
		if err := ValidateSlug(s); !errors.Is(err, ErrInvalidSlug) {
			t.Errorf("ValidateSlug(%q): got %v, want ErrInvalidSlug", s, err)
		}
	}
}
