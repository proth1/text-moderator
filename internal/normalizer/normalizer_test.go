package normalizer

import (
	"testing"
)

func TestNormalize_NFKC(t *testing.T) {
	n := New()
	// Fullwidth A (Ａ U+FF21) should normalize to regular A
	got := n.Normalize("\uFF21\uFF22\uFF23")
	if got != "ABC" {
		t.Errorf("NFKC normalization failed: got %q, want %q", got, "ABC")
	}
}

func TestNormalize_ZeroWidth(t *testing.T) {
	n := New()
	// Zero-width chars inserted between letters
	got := n.Normalize("h\u200Be\u200Dl\u200Blo")
	if got != "hello" {
		t.Errorf("zero-width stripping failed: got %q, want %q", got, "hello")
	}
}

func TestNormalize_Homoglyphs(t *testing.T) {
	n := New()
	// Cyrillic а (U+0430) looks like Latin a
	// Cyrillic е (U+0435) looks like Latin e
	got := n.Normalize("\u0430\u0435")
	if got != "ae" {
		t.Errorf("homoglyph mapping failed: got %q, want %q", got, "ae")
	}
}

func TestNormalize_Leetspeak(t *testing.T) {
	n := New()
	got := n.Normalize("h4t3 sp33ch")
	if got != "hate speech" {
		t.Errorf("leetspeak decode failed: got %q, want %q", got, "hate speech")
	}
}

func TestNormalize_CollapseWhitespace(t *testing.T) {
	n := New()
	got := n.Normalize("  hello   world  ")
	if got != "hello world" {
		t.Errorf("whitespace collapse failed: got %q, want %q", got, "hello world")
	}
}

func TestNormalize_Combined(t *testing.T) {
	n := New()
	// Zero-width + leetspeak + extra spaces
	got := n.Normalize("h\u200B4t3  sp\u200D33ch")
	if got != "hate speech" {
		t.Errorf("combined normalization failed: got %q, want %q", got, "hate speech")
	}
}

func TestNormalize_EmptyString(t *testing.T) {
	n := New()
	got := n.Normalize("")
	if got != "" {
		t.Errorf("empty string normalization failed: got %q, want %q", got, "")
	}
}

func TestNormalize_PlainText(t *testing.T) {
	n := New()
	got := n.Normalize("hello world")
	if got != "hello world" {
		t.Errorf("plain text should be unchanged: got %q, want %q", got, "hello world")
	}
}
