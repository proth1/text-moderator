package classifier

import (
	"context"

	"github.com/proth1/text-moderator/internal/models"
)

// Provider defines the interface for text classification providers.
// Any ML service (HuggingFace, Perspective API, OpenAI, etc.) must implement this.
// Control: MOD-001 (Multi-provider classification orchestration)
type Provider interface {
	// Classify sends text to the provider for classification and returns category scores.
	Classify(ctx context.Context, text string) (*models.CategoryScores, error)

	// Name returns the provider's unique identifier (e.g., "huggingface", "perspective", "openai").
	Name() string

	// ModelInfo returns the model name and version used by this provider.
	ModelInfo() (name string, version string)

	// Health checks if the provider is available.
	Health(ctx context.Context) error
}

// LanguageAwareProvider extends Provider with language-specific classification support.
type LanguageAwareProvider interface {
	Provider

	// ClassifyWithLanguage sends text to the provider with a language hint.
	ClassifyWithLanguage(ctx context.Context, text string, lang string) (*models.CategoryScores, error)

	// SupportedLanguages returns the list of ISO 639-1 language codes supported by this provider.
	SupportedLanguages() []string
}

// ProviderConfig defines routing configuration for a classification provider.
type ProviderConfig struct {
	// Name identifies which provider to use.
	Name string `json:"name" yaml:"name"`

	// Priority determines selection order (lower = higher priority).
	Priority int `json:"priority" yaml:"priority"`

	// Weight for weighted random selection among same-priority providers (0-100).
	Weight int `json:"weight,omitempty" yaml:"weight,omitempty"`

	// Enabled controls whether this provider is active.
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// EnsembleConfig controls ensemble mode where multiple providers run in parallel.
type EnsembleConfig struct {
	Enabled            bool    `json:"enabled" yaml:"enabled"`
	MinProviders       int     `json:"min_providers" yaml:"min_providers"`
	AgreementThreshold float64 `json:"agreement_threshold" yaml:"agreement_threshold"`
	Strategy           string  `json:"strategy" yaml:"strategy"` // "average", "median", "max"
}

// OrchestratorConfig defines configuration for the provider orchestrator.
type OrchestratorConfig struct {
	// Providers lists available providers in priority order.
	Providers []ProviderConfig `json:"providers" yaml:"providers"`

	// FallbackEnabled allows falling back to next provider on failure.
	FallbackEnabled bool `json:"fallback_enabled" yaml:"fallback_enabled"`

	// Ensemble controls parallel multi-provider classification.
	Ensemble *EnsembleConfig `json:"ensemble,omitempty" yaml:"ensemble,omitempty"`
}
