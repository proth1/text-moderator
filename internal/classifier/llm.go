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

// LLMProvider implements the Provider interface using an LLM for classification.
// Designed as a second-pass for ambiguous scores (0.3-0.7 range).
type LLMProvider struct {
	provider   string // "anthropic" or "openai"
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *zap.Logger
}

// LLMConfig holds LLM classification provider configuration.
type LLMConfig struct {
	Provider string // "anthropic" or "openai"
	APIKey   string
	Model    string
	Timeout  time.Duration
}

// NewLLMProvider creates a new LLM-based classification provider.
func NewLLMProvider(cfg LLMConfig, logger *zap.Logger) *LLMProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	model := cfg.Model
	if model == "" {
		if cfg.Provider == "anthropic" {
			model = "claude-sonnet-4-20250514"
		} else {
			model = "gpt-4o-mini"
		}
	}
	return &LLMProvider{
		provider: cfg.Provider,
		apiKey:   cfg.APIKey,
		model:    model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

const classificationPrompt = `You are a content moderation classifier. Analyze the following text and return a JSON object with scores from 0.0 to 1.0 for each category. Higher scores indicate higher confidence of the category being present.

Categories:
- toxicity: Overall toxic content
- hate: Hate speech targeting protected groups
- harassment: Bullying, threats, or intimidation
- sexual_content: Sexually explicit material
- violence: Violent or gory content
- profanity: Vulgar or profane language
- self_harm: Self-harm or suicide related content
- spam: Unsolicited commercial or repetitive content
- pii: Personally identifiable information (names, emails, phones, addresses, SSNs)

Text to analyze:
%s

Respond ONLY with a JSON object, no other text:
{"toxicity":0.0,"hate":0.0,"harassment":0.0,"sexual_content":0.0,"violence":0.0,"profanity":0.0,"self_harm":0.0,"spam":0.0,"pii":0.0}`

func (p *LLMProvider) Classify(ctx context.Context, text string) (*models.CategoryScores, error) {
	prompt := fmt.Sprintf(classificationPrompt, text)

	var body []byte
	var err error
	var apiURL string

	if p.provider == "anthropic" {
		apiURL = "https://api.anthropic.com/v1/messages"
		body, err = p.buildAnthropicRequest(prompt)
	} else {
		apiURL = "https://api.openai.com/v1/chat/completions"
		body, err = p.buildOpenAIRequest(prompt)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to build LLM request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if p.provider == "anthropic" {
		req.Header.Set("x-api-key", p.apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Extract the text content from the response
	content, err := p.extractContent(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to extract LLM response content: %w", err)
	}

	// Parse JSON scores from LLM response
	var scores models.CategoryScores
	if err := json.Unmarshal([]byte(content), &scores); err != nil {
		return nil, fmt.Errorf("failed to parse LLM classification scores: %w (raw: %s)", err, content)
	}

	return &scores, nil
}

func (p *LLMProvider) buildAnthropicRequest(prompt string) ([]byte, error) {
	reqBody := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 256,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	return json.Marshal(reqBody)
}

func (p *LLMProvider) buildOpenAIRequest(prompt string) ([]byte, error) {
	reqBody := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 256,
		"messages": []map[string]interface{}{
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}
	return json.Marshal(reqBody)
}

func (p *LLMProvider) extractContent(body []byte) (string, error) {
	if p.provider == "anthropic" {
		var resp struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return "", err
		}
		if len(resp.Content) == 0 {
			return "", fmt.Errorf("empty response from Anthropic API")
		}
		return resp.Content[0].Text, nil
	}

	// OpenAI format
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI API")
	}
	return resp.Choices[0].Message.Content, nil
}

func (p *LLMProvider) Name() string {
	return "llm-" + p.provider
}

func (p *LLMProvider) ModelInfo() (string, string) {
	return p.model, "latest"
}

func (p *LLMProvider) Health(ctx context.Context) error {
	_, err := p.Classify(ctx, "test")
	return err
}

// IsAmbiguous returns true if any category score falls within the ambiguous range [low, high].
func IsAmbiguous(scores *models.CategoryScores, low, high float64) bool {
	vals := []float64{
		scores.Toxicity, scores.Hate, scores.Harassment,
		scores.SexualContent, scores.Violence, scores.Profanity,
		scores.SelfHarm, scores.Spam, scores.PII,
	}
	for _, v := range vals {
		if v >= low && v <= high {
			return true
		}
	}
	return false
}

// MergeAmbiguousScores replaces only the ambiguous-range categories in primary with LLM scores.
func MergeAmbiguousScores(primary, llm *models.CategoryScores, low, high float64) *models.CategoryScores {
	merged := *primary
	if primary.Toxicity >= low && primary.Toxicity <= high {
		merged.Toxicity = llm.Toxicity
	}
	if primary.Hate >= low && primary.Hate <= high {
		merged.Hate = llm.Hate
	}
	if primary.Harassment >= low && primary.Harassment <= high {
		merged.Harassment = llm.Harassment
	}
	if primary.SexualContent >= low && primary.SexualContent <= high {
		merged.SexualContent = llm.SexualContent
	}
	if primary.Violence >= low && primary.Violence <= high {
		merged.Violence = llm.Violence
	}
	if primary.Profanity >= low && primary.Profanity <= high {
		merged.Profanity = llm.Profanity
	}
	if primary.SelfHarm >= low && primary.SelfHarm <= high {
		merged.SelfHarm = llm.SelfHarm
	}
	if primary.Spam >= low && primary.Spam <= high {
		merged.Spam = llm.Spam
	}
	if primary.PII >= low && primary.PII <= high {
		merged.PII = llm.PII
	}
	return &merged
}
