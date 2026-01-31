package classifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/proth1/text-moderator/internal/models"
	"go.uber.org/zap"
)

// OpenAIProvider implements the Provider interface for OpenAI's Moderation API.
type OpenAIProvider struct {
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// OpenAIConfig holds OpenAI Moderation API configuration.
type OpenAIConfig struct {
	APIKey  string
	Timeout time.Duration
}

// NewOpenAIProvider creates a new OpenAI Moderation classification provider.
func NewOpenAIProvider(cfg OpenAIConfig, logger *zap.Logger) *OpenAIProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &OpenAIProvider{
		apiKey: cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

const openAIModerationURL = "https://api.openai.com/v1/moderations"

type openAIRequest struct {
	Input string `json:"input"`
}

type openAIResponse struct {
	Results []openAIResult `json:"results"`
}

type openAIResult struct {
	CategoryScores openAICategoryScores `json:"category_scores"`
	Flagged        bool                 `json:"flagged"`
}

type openAICategoryScores struct {
	Hate            float64 `json:"hate"`
	HateThreatening float64 `json:"hate/threatening"`
	Harassment      float64 `json:"harassment"`
	SelfHarm        float64 `json:"self-harm"`
	Sexual          float64 `json:"sexual"`
	SexualMinors    float64 `json:"sexual/minors"`
	Violence        float64 `json:"violence"`
	ViolenceGraphic float64 `json:"violence/graphic"`
}

func (p *OpenAIProvider) Classify(ctx context.Context, text string) (*models.CategoryScores, error) {
	reqBody := openAIRequest{Input: text}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIModerationURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai moderation request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai moderation API returned status %d: %s", resp.StatusCode, string(body))
	}

	var oResp openAIResponse
	if err := json.Unmarshal(body, &oResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(oResp.Results) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI moderation API")
	}

	return p.convertScores(&oResp.Results[0]), nil
}

func (p *OpenAIProvider) convertScores(result *openAIResult) *models.CategoryScores {
	cs := result.CategoryScores
	// Map OpenAI categories to our unified CategoryScores.
	// Use the max of related sub-categories.
	toxicity := cs.Hate
	if cs.Harassment > toxicity {
		toxicity = cs.Harassment
	}
	if cs.Violence > toxicity {
		toxicity = cs.Violence
	}

	violence := cs.Violence
	if cs.ViolenceGraphic > violence {
		violence = cs.ViolenceGraphic
	}

	hate := cs.Hate
	if cs.HateThreatening > hate {
		hate = cs.HateThreatening
	}

	sexual := cs.Sexual
	if cs.SexualMinors > sexual {
		sexual = cs.SexualMinors
	}

	return &models.CategoryScores{
		Toxicity:      toxicity,
		Hate:          hate,
		Harassment:    cs.Harassment,
		SexualContent: sexual,
		Violence:      violence,
		Profanity:     cs.Harassment, // OpenAI doesn't have a separate profanity score
		SelfHarm:      cs.SelfHarm,
		Spam:          0.0,
		PII:           0.0,
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) ModelInfo() (string, string) {
	return "openai-moderation", "latest"
}

func (p *OpenAIProvider) Health(ctx context.Context) error {
	_, err := p.Classify(ctx, "test")
	return err
}
