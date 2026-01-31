package normalizer

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Normalizer pre-processes text to defeat Unicode evasion before hashing and classification.
type Normalizer struct {
	homoglyphs map[rune]rune
	leetspeak  map[rune]rune
}

// New creates a Normalizer with default homoglyph and leetspeak mappings.
func New() *Normalizer {
	return &Normalizer{
		homoglyphs: defaultHomoglyphs(),
		leetspeak:  defaultLeetspeak(),
	}
}

// Normalize applies all normalization steps to the input text:
// 1. NFKC Unicode normalization
// 2. Strip zero-width characters
// 3. Homoglyph → Latin mapping
// 4. Leetspeak decoding
// 5. Collapse whitespace
func (n *Normalizer) Normalize(text string) string {
	// Step 1: NFKC normalization (decomposes and recomposes by compatibility)
	text = norm.NFKC.String(text)

	// Step 2: Strip zero-width characters
	text = stripZeroWidth(text)

	// Step 3: Homoglyph → Latin mapping
	text = n.mapRunes(text, n.homoglyphs)

	// Step 4: Leetspeak decoding
	text = n.mapRunes(text, n.leetspeak)

	// Step 5: Collapse whitespace
	text = collapseWhitespace(text)

	return text
}

// stripZeroWidth removes zero-width Unicode characters used to evade filters.
func stripZeroWidth(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		switch r {
		case '\u200B', // zero-width space
			'\u200C', // zero-width non-joiner
			'\u200D', // zero-width joiner
			'\u200E', // left-to-right mark
			'\u200F', // right-to-left mark
			'\u2060', // word joiner
			'\uFEFF': // byte order mark / zero-width no-break space
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// mapRunes replaces runes according to the provided mapping.
func (n *Normalizer) mapRunes(text string, mapping map[rune]rune) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if replacement, ok := mapping[r]; ok {
			b.WriteRune(replacement)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// collapseWhitespace replaces runs of whitespace with a single space and trims edges.
func collapseWhitespace(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	inSpace := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if !inSpace {
				b.WriteRune(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return strings.TrimSpace(b.String())
}
