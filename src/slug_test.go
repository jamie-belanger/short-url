package main

import (
	"testing"
	"log/slog"
	"os"
)

// Test GenerateSlug method.
// WARNING: This may fail when comparing slugs if the first (minimumSlugLength) characters are not unique.
func TestGenerateSlug(t *testing.T) {
	a := &application{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions {
			Level: slog.LevelDebug,
		})),
		links:  make(map[string]string),
	}

	cases := []struct {
		in, want string
	}{
		{"https://www.google.com", "rGu2aeQORKjZ-PDJTfxjc0BJ3PYhmqx38C7flLkWLAk="},
		{"https://go.dev", "bn9Y9rhosoo3vnxlVgnma0Z_8B3sESiOpSRb2oT-6G8="},
		{"https://forecast.weather.gov/MapClick.php?lat=43.6813939&lon=-70.3598961", "juotFN3pYXfmfjrCR0U0O2k3R4d-wG1wow9X0pr-OdQ="},
	}

	for _, c := range cases {
		got := a.generateSlug(c.in)
		if got != c.want[:minimumSlugLength] {
			t.Errorf("GenerateSlug(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
