package classifier

import (
	"math"
	"testing"

	"github.com/proth1/text-moderator/internal/models"
)

func TestAverageValues(t *testing.T) {
	tests := []struct {
		name string
		vals []float64
		want float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{0.5}, 0.5},
		{"multiple", []float64{0.2, 0.4, 0.6}, 0.4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := averageValues(tt.vals)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("averageValues(%v) = %f, want %f", tt.vals, got, tt.want)
			}
		})
	}
}

func TestMedianValues(t *testing.T) {
	tests := []struct {
		name string
		vals []float64
		want float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{0.5}, 0.5},
		{"odd", []float64{0.1, 0.5, 0.9}, 0.5},
		{"even", []float64{0.2, 0.4, 0.6, 0.8}, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := medianValues(tt.vals)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("medianValues(%v) = %f, want %f", tt.vals, got, tt.want)
			}
		})
	}
}

func TestMaxValues(t *testing.T) {
	tests := []struct {
		name string
		vals []float64
		want float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{0.3}, 0.3},
		{"multiple", []float64{0.1, 0.9, 0.5}, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxValues(tt.vals)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("maxValues(%v) = %f, want %f", tt.vals, got, tt.want)
			}
		})
	}
}

func TestCombineScores_Average(t *testing.T) {
	results := []ClassificationResult{
		{Scores: &models.CategoryScores{Toxicity: 0.2, Hate: 0.4}},
		{Scores: &models.CategoryScores{Toxicity: 0.8, Hate: 0.6}},
	}

	combined := combineScores(results, "average")
	if math.Abs(combined.Toxicity-0.5) > 0.001 {
		t.Errorf("average toxicity = %f, want 0.5", combined.Toxicity)
	}
	if math.Abs(combined.Hate-0.5) > 0.001 {
		t.Errorf("average hate = %f, want 0.5", combined.Hate)
	}
}

func TestCombineScores_Max(t *testing.T) {
	results := []ClassificationResult{
		{Scores: &models.CategoryScores{Toxicity: 0.2}},
		{Scores: &models.CategoryScores{Toxicity: 0.8}},
	}

	combined := combineScores(results, "max")
	if math.Abs(combined.Toxicity-0.8) > 0.001 {
		t.Errorf("max toxicity = %f, want 0.8", combined.Toxicity)
	}
}

func TestComputeAgreement_FullAgreement(t *testing.T) {
	results := []ClassificationResult{
		{Scores: &models.CategoryScores{Toxicity: 0.5, Hate: 0.3}},
		{Scores: &models.CategoryScores{Toxicity: 0.5, Hate: 0.3}},
	}

	agreement, disagreed := computeAgreement(results, 0.3)
	if len(disagreed) != 0 {
		t.Errorf("expected no disagreements, got %v", disagreed)
	}
	if agreement["toxicity"] != 1.0 {
		t.Errorf("toxicity agreement = %f, want 1.0", agreement["toxicity"])
	}
}

func TestComputeAgreement_Disagreement(t *testing.T) {
	results := []ClassificationResult{
		{Scores: &models.CategoryScores{Toxicity: 0.1}},
		{Scores: &models.CategoryScores{Toxicity: 0.9}},
	}

	_, disagreed := computeAgreement(results, 0.3)
	found := false
	for _, d := range disagreed {
		if d == "toxicity" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected toxicity disagreement, got %v", disagreed)
	}
}
