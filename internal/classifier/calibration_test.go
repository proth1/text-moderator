package classifier

import (
	"testing"

	"github.com/proth1/text-moderator/internal/models"
)

func TestCalibrate_NoConfig(t *testing.T) {
	var c *Calibrator
	scores := &models.CategoryScores{Toxicity: 0.5}
	result := c.Calibrate("test", scores)
	if result.Toxicity != 0.5 {
		t.Errorf("nil calibrator should return scores unchanged, got %f", result.Toxicity)
	}
}

func TestCalibrate_NoProviderConfig(t *testing.T) {
	c := NewCalibrator(CalibrationConfig{})
	scores := &models.CategoryScores{Toxicity: 0.5}
	result := c.Calibrate("unknown", scores)
	if result.Toxicity != 0.5 {
		t.Errorf("missing provider should return scores unchanged, got %f", result.Toxicity)
	}
}

func TestCalibrate_AppliesOffsetAndScale(t *testing.T) {
	config := CalibrationConfig{
		"test-provider": ProviderCalibration{
			Categories: map[string]CalibrationParams{
				"toxicity": {Offset: -0.1, Scale: 1.2},
			},
		},
	}
	c := NewCalibrator(config)
	scores := &models.CategoryScores{Toxicity: 0.5, Hate: 0.3}

	result := c.Calibrate("test-provider", scores)

	// toxicity: (0.5 + (-0.1)) * 1.2 = 0.4 * 1.2 = 0.48
	expected := 0.48
	if result.Toxicity < expected-0.001 || result.Toxicity > expected+0.001 {
		t.Errorf("calibrated toxicity = %f, want %f", result.Toxicity, expected)
	}

	// hate has no calibration config, should be unchanged
	if result.Hate != 0.3 {
		t.Errorf("hate should be unchanged, got %f", result.Hate)
	}
}

func TestCalibrate_ClampsToRange(t *testing.T) {
	config := CalibrationConfig{
		"test": ProviderCalibration{
			Categories: map[string]CalibrationParams{
				"toxicity": {Offset: 1.0, Scale: 2.0}, // will push above 1.0
				"hate":     {Offset: -2.0, Scale: 1.0}, // will push below 0.0
			},
		},
	}
	c := NewCalibrator(config)
	scores := &models.CategoryScores{Toxicity: 0.5, Hate: 0.5}

	result := c.Calibrate("test", scores)

	if result.Toxicity != 1.0 {
		t.Errorf("toxicity should be clamped to 1.0, got %f", result.Toxicity)
	}
	if result.Hate != 0.0 {
		t.Errorf("hate should be clamped to 0.0, got %f", result.Hate)
	}
}

func TestNewCalibratorFromJSON(t *testing.T) {
	jsonStr := `{"test-provider":{"categories":{"toxicity":{"offset":-0.1,"scale":1.0}}}}`
	c := NewCalibratorFromJSON(jsonStr)
	if c == nil {
		t.Fatal("expected non-nil calibrator from valid JSON")
	}
}

func TestNewCalibratorFromJSON_Empty(t *testing.T) {
	c := NewCalibratorFromJSON("")
	if c != nil {
		t.Error("expected nil calibrator from empty string")
	}
}

func TestNewCalibratorFromJSON_Invalid(t *testing.T) {
	c := NewCalibratorFromJSON("{invalid json")
	if c != nil {
		t.Error("expected nil calibrator from invalid JSON")
	}
}

func TestUpdateFromFeedback(t *testing.T) {
	c := NewCalibrator(CalibrationConfig{})
	c.UpdateFromFeedback("provider-a", "toxicity", -0.05, 1.0)

	scores := &models.CategoryScores{Toxicity: 0.5}
	result := c.Calibrate("provider-a", scores)

	// (0.5 + (-0.05)) * 1.0 = 0.45
	expected := 0.45
	if result.Toxicity < expected-0.001 || result.Toxicity > expected+0.001 {
		t.Errorf("after feedback update, toxicity = %f, want %f", result.Toxicity, expected)
	}
}
