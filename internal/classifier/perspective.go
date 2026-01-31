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

// PerspectiveProvider implements the Provider interface for Google's Perspective API.
type PerspectiveProvider struct {
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// PerspectiveConfig holds Perspective API configuration.
type PerspectiveConfig struct {
	APIKey  string
	Timeout time.Duration
}

// NewPerspectiveProvider creates a new Perspective API classification provider.
func NewPerspectiveProvider(cfg PerspectiveConfig, logger *zap.Logger) *PerspectiveProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &PerspectiveProvider{
		apiKey: cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

const perspectiveURL = "https://commentanalyzer.googleapis.com/v1alpha1/comments:analyze"

type perspectiveRequest struct {
	Comment             perspectiveComment          `json:"comment"`
	RequestedAttributes map[string]json.RawMessage  `json:"requestedAttributes"`
	Languages           []string                    `json:"languages"`
}

type perspectiveComment struct {
	Text string `json:"text"`
}

type perspectiveResponse struct {
	AttributeScores map[string]perspectiveAttribute `json:"attributeScores"`
}

type perspectiveAttribute struct {
	SummaryScore perspectiveScore `json:"summaryScore"`
}

type perspectiveScore struct {
	Value float64 `json:"value"`
}

func (p *PerspectiveProvider) Classify(ctx context.Context, text string) (*models.CategoryScores, error) {
	return p.ClassifyWithLanguage(ctx, text, "en")
}

// ClassifyWithLanguage implements LanguageAwareProvider for Perspective API.
func (p *PerspectiveProvider) ClassifyWithLanguage(ctx context.Context, text string, lang string) (*models.CategoryScores, error) {
	if lang == "" {
		lang = "en"
	}
	reqBody := perspectiveRequest{
		Comment: perspectiveComment{Text: text},
		RequestedAttributes: map[string]json.RawMessage{
			"TOXICITY":            json.RawMessage(`{}`),
			"SEVERE_TOXICITY":     json.RawMessage(`{}`),
			"IDENTITY_ATTACK":     json.RawMessage(`{}`),
			"INSULT":              json.RawMessage(`{}`),
			"PROFANITY":           json.RawMessage(`{}`),
			"THREAT":              json.RawMessage(`{}`),
			"SEXUALLY_EXPLICIT":   json.RawMessage(`{}`),
		},
		Languages: []string{lang},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", perspectiveURL, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perspective API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("perspective API returned status %d: %s", resp.StatusCode, string(body))
	}

	var pResp perspectiveResponse
	if err := json.Unmarshal(body, &pResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return p.convertScores(&pResp), nil
}

func (p *PerspectiveProvider) convertScores(resp *perspectiveResponse) *models.CategoryScores {
	scores := &models.CategoryScores{}

	if attr, ok := resp.AttributeScores["TOXICITY"]; ok {
		scores.Toxicity = attr.SummaryScore.Value
	}
	if attr, ok := resp.AttributeScores["IDENTITY_ATTACK"]; ok {
		scores.Hate = attr.SummaryScore.Value
	}
	if attr, ok := resp.AttributeScores["INSULT"]; ok {
		scores.Harassment = attr.SummaryScore.Value
	}
	if attr, ok := resp.AttributeScores["SEXUALLY_EXPLICIT"]; ok {
		scores.SexualContent = attr.SummaryScore.Value
	}
	if attr, ok := resp.AttributeScores["THREAT"]; ok {
		scores.Violence = attr.SummaryScore.Value
	}
	if attr, ok := resp.AttributeScores["PROFANITY"]; ok {
		scores.Profanity = attr.SummaryScore.Value
	}

	return scores
}

// SupportedLanguages returns the languages supported by Perspective API.
func (p *PerspectiveProvider) SupportedLanguages() []string {
	return []string{"en", "es", "fr", "de", "pt", "it", "ru", "tr", "ar", "zh", "ja", "ko", "hi", "id"}
}

func (p *PerspectiveProvider) Name() string {
	return "perspective"
}

func (p *PerspectiveProvider) ModelInfo() (string, string) {
	return "perspective-api", "v1alpha1"
}

func (p *PerspectiveProvider) Health(ctx context.Context) error {
	// Perspective API doesn't have a dedicated health endpoint;
	// classify a short test string to verify connectivity.
	_, err := p.Classify(ctx, "test")
	return err
}
