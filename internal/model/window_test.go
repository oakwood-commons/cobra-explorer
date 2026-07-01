package model

import "testing"

func TestWindowText_FitsUnchanged(t *testing.T) {
	if got := windowText("short", 0, 10); got != "short" {
		t.Fatalf("expected content unchanged, got %q", got)
	}
}

func TestWindowText_MarkersAndWidth(t *testing.T) {
	s := "0123456789ABCDEF" // 16 runes
	width := 8

	// Scrolled to the start: only a right marker, total width preserved.
	got := windowText(s, 0, width)
	if len([]rune(got)) != width {
		t.Fatalf("start window width = %d, want %d (%q)", len([]rune(got)), width, got)
	}
	if got[len(got)-1] != '>' || got[0] == '<' {
		t.Fatalf("start window should end with '>' and not start with '<': %q", got)
	}

	// Scrolled to the middle: markers on both ends.
	got = windowText(s, 4, width)
	if len([]rune(got)) != width {
		t.Fatalf("middle window width = %d, want %d (%q)", len([]rune(got)), width, got)
	}
	if got[0] != '<' || got[len(got)-1] != '>' {
		t.Fatalf("middle window should have both markers: %q", got)
	}
}

func TestWindowText_ClampsAndShowsEnd(t *testing.T) {
	s := "0123456789ABCDEF" // 16 runes
	width := 8

	// Over-scrolling clamps to the maximum and reveals the last rune.
	got := windowText(s, 1000, width)
	if len([]rune(got)) != width {
		t.Fatalf("end window width = %d, want %d (%q)", len([]rune(got)), width, got)
	}
	if got[0] != '<' {
		t.Fatalf("end window should start with '<': %q", got)
	}
	if got[len(got)-1] != 'F' {
		t.Fatalf("end window should reveal the last rune 'F': %q", got)
	}
}
