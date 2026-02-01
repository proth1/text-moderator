package classifier

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/proth1/text-moderator/internal/models"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// Orchestrator routes classification requests to the appropriate provider
// with fallback chains for resilience.
// Control: MOD-001 (Multi-provider classification orchestration)
type Orchestrator struct {
	providers  map[string]Provider
	breakers   map[string]*gobreaker.CircuitBreaker
	config     OrchestratorConfig
	calibrator *Calibrator
	mu         sync.RWMutex
	logger     *zap.Logger
}

// NewOrchestrator creates a new provider orchestrator.
func NewOrchestrator(cfg OrchestratorConfig, logger *zap.Logger) *Orchestrator {
	return &Orchestrator{
		providers: make(map[string]Provider),
		breakers:  make(map[string]*gobreaker.CircuitBreaker),
		config:    cfg,
		logger:    logger,
	}
}

// RegisterProvider adds a classification provider to the orchestrator.
// A circuit breaker is created per provider: opens after 5 consecutive failures,
// half-opens after 60 seconds.
func (o *Orchestrator) RegisterProvider(provider Provider) {
	o.mu.Lock()
	defer o.mu.Unlock()
	name := provider.Name()
	o.providers[name] = provider

	o.breakers[name] = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "classifier-" + name,
		MaxRequests: 1,                // allow 1 request in half-open state
		Interval:    0,                // don't clear counts in closed state
		Timeout:     60 * time.Second, // half-open after 60s
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			o.logger.Warn("circuit breaker state change",
				zap.String("breaker", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	})

	o.logger.Info("registered classification provider", zap.String("provider", name))
}

// Classify routes the classification request to the highest-priority available provider,
// falling back to alternatives on failure when fallback is enabled.
func (o *Orchestrator) Classify(ctx context.Context, text string) (*ClassificationResult, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	ordered := o.orderedProviders()
	if len(ordered) == 0 {
		return nil, fmt.Errorf("no classification providers registered")
	}

	var lastErr error
	for _, pcfg := range ordered {
		provider, exists := o.providers[pcfg.Name]
		if !exists {
			o.logger.Warn("configured provider not registered", zap.String("provider", pcfg.Name))
			continue
		}

		breaker := o.breakers[pcfg.Name]
		var scores *models.CategoryScores

		if breaker != nil {
			result, err := breaker.Execute(func() (interface{}, error) {
				return provider.Classify(ctx, text)
			})
			if err != nil {
				lastErr = err
				o.logger.Warn("provider classification failed",
					zap.String("provider", pcfg.Name),
					zap.Error(err),
				)
				if o.config.FallbackEnabled {
					continue
				}
				return nil, fmt.Errorf("provider %s failed: %w", pcfg.Name, err)
			}
			scores = result.(*models.CategoryScores)
		} else {
			var err error
			scores, err = provider.Classify(ctx, text)
			if err != nil {
				lastErr = err
				o.logger.Warn("provider classification failed",
					zap.String("provider", pcfg.Name),
					zap.Error(err),
				)
				if o.config.FallbackEnabled {
					continue
				}
				return nil, fmt.Errorf("provider %s failed: %w", pcfg.Name, err)
			}
		}

		if o.calibrator != nil {
			scores = o.calibrator.Calibrate(pcfg.Name, scores)
		}

		modelName, modelVersion := provider.ModelInfo()
		return &ClassificationResult{
			Scores:       scores,
			ProviderName: provider.Name(),
			ModelName:    modelName,
			ModelVersion: modelVersion,
		}, nil
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

// ClassifyWithProvider routes classification to a specific named provider.
func (o *Orchestrator) ClassifyWithProvider(ctx context.Context, text string, providerName string) (*ClassificationResult, error) {
	o.mu.RLock()
	provider, exists := o.providers[providerName]
	o.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider %q not registered", providerName)
	}

	scores, err := provider.Classify(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("provider %s failed: %w", providerName, err)
	}

	modelName, modelVersion := provider.ModelInfo()
	return &ClassificationResult{
		Scores:       scores,
		ProviderName: provider.Name(),
		ModelName:    modelName,
		ModelVersion: modelVersion,
	}, nil
}

// SetCalibrator sets an optional calibrator for score normalization.
func (o *Orchestrator) SetCalibrator(c *Calibrator) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.calibrator = c
}

// SetEnsembleConfig sets the ensemble configuration.
func (o *Orchestrator) SetEnsembleConfig(cfg *EnsembleConfig) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.config.Ensemble = cfg
}

// IsEnsembleEnabled returns whether ensemble mode is configured and enabled.
func (o *Orchestrator) IsEnsembleEnabled() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.config.Ensemble != nil && o.config.Ensemble.Enabled
}

// ClassifyWithLanguage routes classification with a language hint.
// Prefers LanguageAwareProvider implementations for non-English text.
func (o *Orchestrator) ClassifyWithLanguage(ctx context.Context, text string, lang string) (*ClassificationResult, error) {
	// For English or empty language, use standard classification
	if lang == "" || lang == "en" {
		result, err := o.Classify(ctx, text)
		if result != nil {
			result.DetectedLanguage = lang
		}
		return result, err
	}

	o.mu.RLock()
	ordered := o.orderedProviders()
	providers := make(map[string]Provider, len(o.providers))
	for k, v := range o.providers {
		providers[k] = v
	}
	calibrator := o.calibrator
	fallbackEnabled := o.config.FallbackEnabled
	o.mu.RUnlock()

	if len(ordered) == 0 {
		return nil, fmt.Errorf("no classification providers registered")
	}

	// Try language-aware providers first
	var lastErr error
	for _, pcfg := range ordered {
		provider, exists := providers[pcfg.Name]
		if !exists {
			continue
		}

		langProvider, isLangAware := provider.(LanguageAwareProvider)
		if !isLangAware {
			continue
		}

		// Check if this provider supports the detected language
		supported := false
		for _, sl := range langProvider.SupportedLanguages() {
			if sl == lang {
				supported = true
				break
			}
		}
		if !supported {
			continue
		}

		scores, err := langProvider.ClassifyWithLanguage(ctx, text, lang)
		if err != nil {
			lastErr = err
			o.logger.Warn("language-aware provider failed",
				zap.String("provider", pcfg.Name),
				zap.String("language", lang),
				zap.Error(err),
			)
			if fallbackEnabled {
				continue
			}
			return nil, fmt.Errorf("provider %s failed: %w", pcfg.Name, err)
		}

		if calibrator != nil {
			scores = calibrator.Calibrate(pcfg.Name, scores)
		}

		modelName, modelVersion := provider.ModelInfo()
		return &ClassificationResult{
			Scores:           scores,
			ProviderName:     provider.Name(),
			ModelName:        modelName,
			ModelVersion:     modelVersion,
			DetectedLanguage: lang,
		}, nil
	}

	// Fall back to standard classification if no language-aware provider succeeded
	if lastErr != nil {
		o.logger.Warn("no language-aware provider succeeded, falling back to standard classification",
			zap.String("language", lang),
			zap.Error(lastErr),
		)
	}

	result, err := o.Classify(ctx, text)
	if result != nil {
		result.DetectedLanguage = lang
	}
	return result, err
}

// HealthCheck returns the health status of all registered providers.
func (o *Orchestrator) HealthCheck(ctx context.Context) map[string]error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	results := make(map[string]error)
	for name, provider := range o.providers {
		results[name] = provider.Health(ctx)
	}
	return results
}

// orderedProviders returns enabled providers sorted by priority (ascending).
func (o *Orchestrator) orderedProviders() []ProviderConfig {
	var enabled []ProviderConfig
	for _, p := range o.config.Providers {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	sort.Slice(enabled, func(i, j int) bool {
		return enabled[i].Priority < enabled[j].Priority
	})
	return enabled
}

// ClassificationResult holds the result from a provider classification.
type ClassificationResult struct {
	Scores           *models.CategoryScores
	ProviderName     string
	ModelName        string
	ModelVersion     string
	DetectedLanguage string
}
