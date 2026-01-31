package langdetect

import (
	"strings"

	"github.com/pemistahl/lingua-go"
)

// DetectionResult holds the result of language detection.
type DetectionResult struct {
	Language   string  `json:"language"`
	Confidence float64 `json:"confidence"`
}

// Detector wraps lingua-go for language detection.
type Detector struct {
	detector lingua.LanguageDetector
}

// New creates a Detector supporting all lingua languages.
func New() *Detector {
	detector := lingua.NewLanguageDetectorBuilder().
		FromAllLanguages().
		WithMinimumRelativeDistance(0.25).
		Build()

	return &Detector{detector: detector}
}

// NewWithLanguages creates a Detector supporting only the specified languages.
func NewWithLanguages(languages []lingua.Language) *Detector {
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		WithMinimumRelativeDistance(0.25).
		Build()

	return &Detector{detector: detector}
}

// Detect identifies the language of the input text.
// Returns "en" as default if detection fails or confidence is too low.
func (d *Detector) Detect(text string) DetectionResult {
	if len(text) < 10 {
		return DetectionResult{Language: "en", Confidence: 0.0}
	}

	language, exists := d.detector.DetectLanguageOf(text)
	if !exists {
		return DetectionResult{Language: "en", Confidence: 0.0}
	}

	confidences := d.detector.ComputeLanguageConfidenceValues(text)
	var confidence float64
	for _, cv := range confidences {
		if cv.Language() == language {
			confidence = cv.Value()
			break
		}
	}

	return DetectionResult{
		Language:   strings.ToLower(language.IsoCode639_1().String()),
		Confidence: confidence,
	}
}
