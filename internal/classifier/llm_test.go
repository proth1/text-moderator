package classifier

import (
	"testing"

	"github.com/proth1/text-moderator/internal/models"
)

func TestIsAmbiguous(t *testing.T) {
	tests := []struct {
		name   string
		scores *models.CategoryScores
		want   bool
	}{
		{
			name:   "clear scores",
			scores: &models.CategoryScores{Toxicity: 0.1, Hate: 0.9},
			want:   false,
		},
		{
			name:   "ambiguous toxicity",
			scores: &models.CategoryScores{Toxicity: 0.5},
			want:   true,
		},
		{
			name:   "boundary low",
			scores: &models.CategoryScores{Hate: 0.3},
			want:   true,
		},
		{
			name:   "boundary high",
			scores: &models.CategoryScores{Violence: 0.7},
			want:   true,
		},
		{
			name:   "just below range",
			scores: &models.CategoryScores{Toxicity: 0.29},
			want:   false,
		},
		{
			name:   "just above range",
			scores: &models.CategoryScores{Toxicity: 0.71},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAmbiguous(tt.scores, 0.3, 0.7)
			if got != tt.want {
				t.Errorf("IsAmbiguous() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeAmbiguousScores(t *testing.T) {
	primary := &models.CategoryScores{
		Toxicity: 0.5, // ambiguous
		Hate:     0.1, // clear
		Violence: 0.9, // clear
		Spam:     0.4, // ambiguous
	}
	llm := &models.CategoryScores{
		Toxicity: 0.8,
		Hate:     0.2,
		Violence: 0.3,
		Spam:     0.7,
	}

	merged := MergeAmbiguousScores(primary, llm, 0.3, 0.7)

	// Ambiguous categories should use LLM scores
	if merged.Toxicity != 0.8 {
		t.Errorf("ambiguous toxicity should use LLM score 0.8, got %f", merged.Toxicity)
	}
	if merged.Spam != 0.7 {
		t.Errorf("ambiguous spam should use LLM score 0.7, got %f", merged.Spam)
	}

	// Clear categories should keep primary scores
	if merged.Hate != 0.1 {
		t.Errorf("clear hate should keep primary score 0.1, got %f", merged.Hate)
	}
	if merged.Violence != 0.9 {
		t.Errorf("clear violence should keep primary score 0.9, got %f", merged.Violence)
	}
}
