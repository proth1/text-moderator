package fairness

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/proth1/text-moderator/internal/models"
	"go.uber.org/zap"
)

// BiasDetector analyzes moderation decisions for demographic bias
// by comparing action rates across different content sources and languages.
type BiasDetector struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewBiasDetector creates a new bias detection pipeline.
func NewBiasDetector(db *pgxpool.Pool, logger *zap.Logger) *BiasDetector {
	return &BiasDetector{db: db, logger: logger}
}

// FairnessReport contains bias metrics for a given time window.
type FairnessReport struct {
	Window         time.Duration          `json:"window"`
	GeneratedAt    time.Time              `json:"generated_at"`
	TotalDecisions int                    `json:"total_decisions"`
	ActionRates    map[string]float64     `json:"action_rates"`
	BySource       map[string]SourceStats `json:"by_source"`
	ByLanguage     map[string]SourceStats `json:"by_language"`
	BiasAlerts     []BiasAlert            `json:"bias_alerts,omitempty"`
}

// SourceStats holds action rates for a specific source/language segment.
type SourceStats struct {
	Count       int                `json:"count"`
	ActionRates map[string]float64 `json:"action_rates"`
}

// BiasAlert is raised when a segment's action rate deviates significantly
// from the overall baseline.
type BiasAlert struct {
	Segment    string  `json:"segment"`
	Dimension  string  `json:"dimension"` // "source" or "language"
	Action     string  `json:"action"`
	Rate       float64 `json:"rate"`
	Baseline   float64 `json:"baseline"`
	Deviation  float64 `json:"deviation"` // standard deviations from baseline
	Severity   string  `json:"severity"`  // "warning" or "critical"
}

// GenerateReport analyzes the last `window` of moderation decisions for bias.
func (d *BiasDetector) GenerateReport(ctx context.Context, window time.Duration) (*FairnessReport, error) {
	since := time.Now().Add(-window)

	report := &FairnessReport{
		Window:      window,
		GeneratedAt: time.Now(),
		ActionRates: make(map[string]float64),
		BySource:    make(map[string]SourceStats),
		ByLanguage:  make(map[string]SourceStats),
	}

	// Overall action rates
	rows, err := d.db.Query(ctx,
		`SELECT automated_action, COUNT(*) FROM moderation_decisions
		 WHERE created_at >= $1 GROUP BY automated_action`, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query overall rates: %w", err)
	}
	defer rows.Close()

	actionCounts := make(map[string]int)
	total := 0
	for rows.Next() {
		var action models.PolicyAction
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			continue
		}
		actionCounts[string(action)] = count
		total += count
	}
	report.TotalDecisions = total

	if total == 0 {
		return report, nil
	}

	for action, count := range actionCounts {
		report.ActionRates[action] = float64(count) / float64(total)
	}

	// Per-source action rates
	sourceRows, err := d.db.Query(ctx,
		`SELECT COALESCE(s.source, 'unknown'), d.automated_action, COUNT(*)
		 FROM moderation_decisions d
		 JOIN text_submissions s ON s.id = d.submission_id
		 WHERE d.created_at >= $1
		 GROUP BY s.source, d.automated_action`, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query source rates: %w", err)
	}
	defer sourceRows.Close()

	sourceActionCounts := make(map[string]map[string]int)
	sourceTotals := make(map[string]int)
	for sourceRows.Next() {
		var source string
		var action models.PolicyAction
		var count int
		if err := sourceRows.Scan(&source, &action, &count); err != nil {
			continue
		}
		if sourceActionCounts[source] == nil {
			sourceActionCounts[source] = make(map[string]int)
		}
		sourceActionCounts[source][string(action)] = count
		sourceTotals[source] += count
	}

	for source, counts := range sourceActionCounts {
		stats := SourceStats{
			Count:       sourceTotals[source],
			ActionRates: make(map[string]float64),
		}
		for action, count := range counts {
			stats.ActionRates[action] = float64(count) / float64(sourceTotals[source])
		}
		report.BySource[source] = stats
	}

	// Detect bias: flag segments where block rate deviates >2 standard deviations
	baselineBlockRate := report.ActionRates["block"]
	for source, stats := range report.BySource {
		blockRate := stats.ActionRates["block"]
		if stats.Count < 10 { // Skip small samples
			continue
		}
		// Standard error of proportion
		se := math.Sqrt(baselineBlockRate * (1 - baselineBlockRate) / float64(stats.Count))
		if se == 0 {
			continue
		}
		zScore := (blockRate - baselineBlockRate) / se
		if math.Abs(zScore) > 2 {
			severity := "warning"
			if math.Abs(zScore) > 3 {
				severity = "critical"
			}
			report.BiasAlerts = append(report.BiasAlerts, BiasAlert{
				Segment:   source,
				Dimension: "source",
				Action:    "block",
				Rate:      blockRate,
				Baseline:  baselineBlockRate,
				Deviation: zScore,
				Severity:  severity,
			})
		}
	}

	if len(report.BiasAlerts) > 0 {
		d.logger.Warn("bias alerts detected",
			zap.Int("alert_count", len(report.BiasAlerts)),
			zap.Int("total_decisions", total),
		)
	}

	return report, nil
}
