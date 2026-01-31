package langdetect

import (
	"testing"
)

func TestDetect_ShortText(t *testing.T) {
	d := New()
	result := d.Detect("hi")
	if result.Language != "en" {
		t.Errorf("short text should default to en, got %q", result.Language)
	}
	if result.Confidence != 0.0 {
		t.Errorf("short text confidence should be 0.0, got %f", result.Confidence)
	}
}

func TestDetect_EmptyText(t *testing.T) {
	d := New()
	result := d.Detect("")
	if result.Language != "en" {
		t.Errorf("empty text should default to en, got %q", result.Language)
	}
}

func TestDetect_EnglishText(t *testing.T) {
	d := New()
	result := d.Detect("This is a longer sentence in English that should be detected correctly")
	if result.Language != "en" {
		t.Errorf("expected en, got %q", result.Language)
	}
	if result.Confidence <= 0.0 {
		t.Errorf("confidence should be positive, got %f", result.Confidence)
	}
}

func TestDetect_SpanishText(t *testing.T) {
	d := New()
	result := d.Detect("Esta es una oración larga en español que debería ser detectada correctamente")
	if result.Language != "es" {
		t.Errorf("expected es, got %q", result.Language)
	}
}

func TestDetect_FrenchText(t *testing.T) {
	d := New()
	result := d.Detect("Ceci est une longue phrase en français qui devrait être détectée correctement")
	if result.Language != "fr" {
		t.Errorf("expected fr, got %q", result.Language)
	}
}
