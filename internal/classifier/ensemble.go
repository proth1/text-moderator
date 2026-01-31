package classifier

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/proth1/text-moderator/internal/models"
	"go.uber.org/zap"
)

// EnsembleResult holds the combined result from multiple providers.
type EnsembleResult struct {
	CombinedScores     *models.CategoryScores
	ProviderResults    []ClassificationResult
	AgreementScores    map[string]float64 // per-category agreement (0-1)
	HasDisagreement    bool
	DisagreedCategories []string
}

// ClassifyEnsemble runs multiple providers in parallel and combines results.
func (o *Orchestrator) ClassifyEnsemble(ctx context.Context, text string) (*EnsembleResult, error) {
	o.mu.RLock()
	ordered := o.orderedProviders()
	providers := make(map[string]Provider, len(o.providers))
	for k, v := range o.providers {
		providers[k] = v
	}
	calibrator := o.calibrator
	ensembleCfg := o.config.Ensemble
	o.mu.RUnlock()

	if ensembleCfg == nil {
		return nil, fmt.Errorf("ensemble mode not configured")
	}

	if len(ordered) < ensembleCfg.MinProviders {
		return nil, fmt.Errorf("need at least %d providers for ensemble, have %d", ensembleCfg.MinProviders, len(ordered))
	}

	type providerResult struct {
		result ClassificationResult
		err    error
		index  int
	}

	results := make(chan providerResult, len(ordered))
	var wg sync.WaitGroup

	for i, pcfg := range ordered {
		provider, exists := providers[pcfg.Name]
		if !exists {
			continue
		}

		wg.Add(1)
		go func(idx int, p Provider, name string) {
			defer wg.Done()
			scores, err := p.Classify(ctx, text)
			if err != nil {
				results <- providerResult{err: err, index: idx}
				return
			}
			if calibrator != nil {
				scores = calibrator.Calibrate(name, scores)
			}
			modelName, modelVersion := p.ModelInfo()
			results <- providerResult{
				result: ClassificationResult{
					Scores:       scores,
					ProviderName: name,
					ModelName:    modelName,
					ModelVersion: modelVersion,
				},
				index: idx,
			}
		}(i, provider, pcfg.Name)
	}

	// Close results channel after all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	var successful []ClassificationResult
	for pr := range results {
		if pr.err != nil {
			o.logger.Warn("ensemble provider failed",
				zap.Error(pr.err),
			)
			continue
		}
		successful = append(successful, pr.result)
	}

	if len(successful) < ensembleCfg.MinProviders {
		return nil, fmt.Errorf("only %d providers succeeded, need %d", len(successful), ensembleCfg.MinProviders)
	}

	// Combine scores based on strategy
	combined := combineScores(successful, ensembleCfg.Strategy)

	// Compute agreement scores
	agreementThreshold := ensembleCfg.AgreementThreshold
	if agreementThreshold == 0 {
		agreementThreshold = 0.3
	}
	agreement, disagreedCategories := computeAgreement(successful, agreementThreshold)

	return &EnsembleResult{
		CombinedScores:     combined,
		ProviderResults:    successful,
		AgreementScores:    agreement,
		HasDisagreement:    len(disagreedCategories) > 0,
		DisagreedCategories: disagreedCategories,
	}, nil
}

func combineScores(results []ClassificationResult, strategy string) *models.CategoryScores {
	if len(results) == 0 {
		return &models.CategoryScores{}
	}

	categories := []string{"toxicity", "hate", "harassment", "sexual_content", "violence", "profanity", "self_harm", "spam", "pii"}
	categoryValues := make(map[string][]float64)

	for _, cat := range categories {
		categoryValues[cat] = make([]float64, 0, len(results))
	}

	for _, r := range results {
		categoryValues["toxicity"] = append(categoryValues["toxicity"], r.Scores.Toxicity)
		categoryValues["hate"] = append(categoryValues["hate"], r.Scores.Hate)
		categoryValues["harassment"] = append(categoryValues["harassment"], r.Scores.Harassment)
		categoryValues["sexual_content"] = append(categoryValues["sexual_content"], r.Scores.SexualContent)
		categoryValues["violence"] = append(categoryValues["violence"], r.Scores.Violence)
		categoryValues["profanity"] = append(categoryValues["profanity"], r.Scores.Profanity)
		categoryValues["self_harm"] = append(categoryValues["self_harm"], r.Scores.SelfHarm)
		categoryValues["spam"] = append(categoryValues["spam"], r.Scores.Spam)
		categoryValues["pii"] = append(categoryValues["pii"], r.Scores.PII)
	}

	combine := averageValues
	switch strategy {
	case "median":
		combine = medianValues
	case "max":
		combine = maxValues
	}

	return &models.CategoryScores{
		Toxicity:      combine(categoryValues["toxicity"]),
		Hate:          combine(categoryValues["hate"]),
		Harassment:    combine(categoryValues["harassment"]),
		SexualContent: combine(categoryValues["sexual_content"]),
		Violence:      combine(categoryValues["violence"]),
		Profanity:     combine(categoryValues["profanity"]),
		SelfHarm:      combine(categoryValues["self_harm"]),
		Spam:          combine(categoryValues["spam"]),
		PII:           combine(categoryValues["pii"]),
	}
}

func averageValues(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func medianValues(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func maxValues(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// computeAgreement calculates per-category agreement scores and identifies disagreements.
// Agreement is 1.0 - (maxSpread / threshold). Categories where spread exceeds threshold are disagreements.
func computeAgreement(results []ClassificationResult, threshold float64) (map[string]float64, []string) {
	type catExtractor func(*models.CategoryScores) float64

	extractors := map[string]catExtractor{
		"toxicity":       func(s *models.CategoryScores) float64 { return s.Toxicity },
		"hate":           func(s *models.CategoryScores) float64 { return s.Hate },
		"harassment":     func(s *models.CategoryScores) float64 { return s.Harassment },
		"sexual_content": func(s *models.CategoryScores) float64 { return s.SexualContent },
		"violence":       func(s *models.CategoryScores) float64 { return s.Violence },
		"profanity":      func(s *models.CategoryScores) float64 { return s.Profanity },
		"self_harm":      func(s *models.CategoryScores) float64 { return s.SelfHarm },
		"spam":           func(s *models.CategoryScores) float64 { return s.Spam },
		"pii":            func(s *models.CategoryScores) float64 { return s.PII },
	}

	agreement := make(map[string]float64)
	var disagreed []string

	for cat, extract := range extractors {
		min, max := math.MaxFloat64, -math.MaxFloat64
		for _, r := range results {
			val := extract(r.Scores)
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
		spread := max - min
		agr := 1.0 - (spread / threshold)
		if agr < 0 {
			agr = 0
		}
		agreement[cat] = agr
		if spread > threshold {
			disagreed = append(disagreed, cat)
		}
	}

	sort.Strings(disagreed)
	return agreement, disagreed
}
