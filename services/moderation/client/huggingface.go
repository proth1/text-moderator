package client

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

// Control: MOD-001 (HuggingFace API integration for text classification)

// HuggingFaceClient handles interactions with HuggingFace Inference API
type HuggingFaceClient struct {
	apiKey     string
	modelURL   string
	httpClient *http.Client
	logger     *zap.Logger
}

// Config holds HuggingFace client configuration
type Config struct {
	APIKey  string
	ModelURL string
	Timeout time.Duration
}

// HuggingFaceRequest represents the request to HuggingFace API
type HuggingFaceRequest struct {
	Inputs     string                 `json:"inputs"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// HuggingFaceResponse represents the response from HuggingFace API
// The toxic-bert model returns multi-label classification results
type HuggingFaceResponse [][]LabelScore

// LabelScore represents a single label and its confidence score
type LabelScore struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

// NewHuggingFaceClient creates a new HuggingFace API client
func NewHuggingFaceClient(cfg Config, logger *zap.Logger) *HuggingFaceClient {
	return &HuggingFaceClient{
		apiKey:   cfg.APIKey,
		modelURL: cfg.ModelURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

// ClassifyText sends text to HuggingFace for toxicity classification
func (c *HuggingFaceClient) ClassifyText(ctx context.Context, text string) (*models.CategoryScores, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Prepare request
	reqBody := HuggingFaceRequest{
		Inputs: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.modelURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// Log request
	c.logger.Debug("sending request to HuggingFace",
		zap.String("model_url", c.modelURL),
		zap.Int("text_length", len(text)),
	)

	// Send request with retry logic
	var resp *http.Response
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode != http.StatusServiceUnavailable {
			break
		}

		if attempt < maxRetries-1 {
			waitTime := time.Duration(attempt+1) * time.Second
			c.logger.Warn("retrying HuggingFace request",
				zap.Int("attempt", attempt+1),
				zap.Duration("wait_time", waitTime),
			)
			time.Sleep(waitTime)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("HuggingFace API error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(body)),
		)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var hfResp HuggingFaceResponse
	if err := json.Unmarshal(body, &hfResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert HuggingFace response to CategoryScores
	scores, err := c.convertToScores(hfResp)
	if err != nil {
		return nil, fmt.Errorf("failed to convert scores: %w", err)
	}

	c.logger.Debug("received classification results",
		zap.Float64("toxicity", scores.Toxicity),
		zap.Float64("hate", scores.Hate),
		zap.Float64("harassment", scores.Harassment),
	)

	return scores, nil
}

// convertToScores converts HuggingFace response to CategoryScores
func (c *HuggingFaceClient) convertToScores(resp HuggingFaceResponse) (*models.CategoryScores, error) {
	if len(resp) == 0 || len(resp[0]) == 0 {
		return nil, fmt.Errorf("empty response from HuggingFace")
	}

	scores := &models.CategoryScores{}

	// Map labels to category scores
	// The toxic-bert model uses specific label names
	for _, labelScore := range resp[0] {
		switch labelScore.Label {
		case "toxic", "toxicity":
			scores.Toxicity = labelScore.Score
		case "severe_toxic", "severe_toxicity":
			// Use max of severe and regular toxicity
			if labelScore.Score > scores.Toxicity {
				scores.Toxicity = labelScore.Score
			}
		case "obscene", "profanity":
			scores.Profanity = labelScore.Score
		case "threat", "violence":
			scores.Violence = labelScore.Score
		case "insult", "harassment":
			scores.Harassment = labelScore.Score
		case "identity_hate", "hate":
			scores.Hate = labelScore.Score
		case "sexual_explicit", "sexual":
			scores.SexualContent = labelScore.Score
		}
	}

	return scores, nil
}

// Health checks the HuggingFace API health
func (c *HuggingFaceClient) Health(ctx context.Context) error {
	// Simple test classification to verify API is accessible
	testText := "Hello, this is a test."

	req, err := http.NewRequestWithContext(ctx, "POST", c.modelURL, bytes.NewBufferString(`{"inputs":"test"}`))
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HuggingFace API health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("HuggingFace API returned status %d", resp.StatusCode)
	}

	c.logger.Debug("HuggingFace health check passed", zap.String("test_text", testText))
	return nil
}
