package classifier

import (
	"context"
	"encoding/json"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/proth1/text-moderator/internal/models"
)

// CalibrationParams holds the offset and scale for a single category.
type CalibrationParams struct {
	Offset float64 `json:"offset"`
	Scale  float64 `json:"scale"`
}

// ProviderCalibration holds calibration parameters per category for a provider.
type ProviderCalibration struct {
	Categories map[string]CalibrationParams `json:"categories"`
}

// CalibrationConfig maps provider names to their calibration parameters.
type CalibrationConfig map[string]ProviderCalibration

// Calibrator normalizes raw provider scores to a unified 0-1 scale.
type Calibrator struct {
	config CalibrationConfig
}

// NewCalibrator creates a Calibrator from the given config.
func NewCalibrator(config CalibrationConfig) *Calibrator {
	return &Calibrator{config: config}
}

// NewCalibratorFromJSON creates a Calibrator from a JSON string.
// Returns nil if the JSON is empty or invalid.
func NewCalibratorFromJSON(configJSON string) *Calibrator {
	if configJSON == "" {
		return nil
	}
	var config CalibrationConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil
	}
	return NewCalibrator(config)
}

// Calibrate applies calibration to raw scores from a named provider.
// Formula: calibrated = clamp((raw + offset) * scale, 0.0, 1.0)
// If no calibration exists for the provider, scores are returned unchanged.
func (c *Calibrator) Calibrate(providerName string, scores *models.CategoryScores) *models.CategoryScores {
	if c == nil || c.config == nil {
		return scores
	}

	providerCal, exists := c.config[providerName]
	if !exists {
		return scores
	}

	calibrated := *scores // copy

	calibrated.Toxicity = c.calibrateValue(providerCal, "toxicity", scores.Toxicity)
	calibrated.Hate = c.calibrateValue(providerCal, "hate", scores.Hate)
	calibrated.Harassment = c.calibrateValue(providerCal, "harassment", scores.Harassment)
	calibrated.SexualContent = c.calibrateValue(providerCal, "sexual_content", scores.SexualContent)
	calibrated.Violence = c.calibrateValue(providerCal, "violence", scores.Violence)
	calibrated.Profanity = c.calibrateValue(providerCal, "profanity", scores.Profanity)
	calibrated.SelfHarm = c.calibrateValue(providerCal, "self_harm", scores.SelfHarm)
	calibrated.Spam = c.calibrateValue(providerCal, "spam", scores.Spam)
	calibrated.PII = c.calibrateValue(providerCal, "pii", scores.PII)

	return &calibrated
}

func (c *Calibrator) calibrateValue(providerCal ProviderCalibration, category string, raw float64) float64 {
	params, exists := providerCal.Categories[category]
	if !exists {
		return raw
	}
	return clamp((raw+params.Offset)*params.Scale, 0.0, 1.0)
}

// UpdateFromFeedback dynamically updates calibration offsets from feedback data.
func (c *Calibrator) UpdateFromFeedback(providerName string, category string, offset float64, scale float64) {
	if c.config == nil {
		c.config = make(CalibrationConfig)
	}
	providerCal, exists := c.config[providerName]
	if !exists {
		providerCal = ProviderCalibration{Categories: make(map[string]CalibrationParams)}
	}
	providerCal.Categories[category] = CalibrationParams{Offset: offset, Scale: scale}
	c.config[providerName] = providerCal
}

// LoadFromDB computes dynamic calibration offsets from feedback data.
// For providers with high disagree rates, it adjusts offsets to shift scores
// toward agreement with human reviewers.
func (c *Calibrator) LoadFromDB(ctx context.Context, db *pgxpool.Pool) error {
	query := `
		SELECT
			provider_name,
			review_outcome,
			AVG((category_scores->>'toxicity')::float) AS avg_toxicity,
			AVG((category_scores->>'hate')::float) AS avg_hate,
			AVG((category_scores->>'harassment')::float) AS avg_harassment,
			COUNT(*) AS cnt
		FROM calibration_data
		WHERE created_at >= NOW() - INTERVAL '30 days'
		GROUP BY provider_name, review_outcome
		HAVING COUNT(*) >= 10
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	type feedbackRow struct {
		providerName string
		outcome      string
		avgToxicity  float64
		avgHate      float64
		avgHarassment float64
		count        int
	}

	var feedbackRows []feedbackRow
	for rows.Next() {
		var r feedbackRow
		if err := rows.Scan(&r.providerName, &r.outcome, &r.avgToxicity, &r.avgHate, &r.avgHarassment, &r.count); err != nil {
			continue
		}
		feedbackRows = append(feedbackRows, r)
	}

	// Compute offsets: if disagree cases have high scores, provider is too sensitive (negative offset)
	// if disagree cases have low scores, provider is missing things (positive offset)
	for _, r := range feedbackRows {
		if r.outcome != "disagree" || r.count < 10 {
			continue
		}
		// If disagree avg score > 0.5, provider is over-flagging → negative offset
		// If disagree avg score < 0.3, provider is under-flagging → positive offset
		toxOffset := 0.0
		if r.avgToxicity > 0.5 {
			toxOffset = -(r.avgToxicity - 0.5) * 0.5
		} else if r.avgToxicity < 0.3 {
			toxOffset = (0.3 - r.avgToxicity) * 0.5
		}
		if math.Abs(toxOffset) > 0.01 {
			c.UpdateFromFeedback(r.providerName, "toxicity", toxOffset, 1.0)
		}
	}

	return nil
}

func clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(max, val))
}
