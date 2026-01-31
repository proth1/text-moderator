package classifier

import (
	"context"

	"github.com/proth1/text-moderator/internal/models"
	"github.com/proth1/text-moderator/services/moderation/client"
)

// HuggingFaceProvider adapts the existing HuggingFace client to the Provider interface.
type HuggingFaceProvider struct {
	client *client.HuggingFaceClient
}

// NewHuggingFaceProvider creates a new HuggingFace classification provider.
func NewHuggingFaceProvider(hfClient *client.HuggingFaceClient) *HuggingFaceProvider {
	return &HuggingFaceProvider{client: hfClient}
}

func (p *HuggingFaceProvider) Classify(ctx context.Context, text string) (*models.CategoryScores, error) {
	return p.client.ClassifyText(ctx, text)
}

func (p *HuggingFaceProvider) Name() string {
	return "huggingface"
}

func (p *HuggingFaceProvider) ModelInfo() (string, string) {
	return "s-nlp/roberta_toxicity_classifier", "v1"
}

func (p *HuggingFaceProvider) Health(ctx context.Context) error {
	return p.client.Health(ctx)
}
