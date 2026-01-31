package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Reporter generates compliance reports from evidence records and system data.
// Supports SOC 2 Type II, ISO 42001, EU AI Act, and GDPR audit formats.
// Control: AUD-001 (Compliance reporting and evidence export)
type Reporter struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewReporter creates a new compliance reporter.
func NewReporter(db *pgxpool.Pool, logger *zap.Logger) *Reporter {
	return &Reporter{db: db, logger: logger}
}

// ReportType identifies the compliance framework for report generation.
type ReportType string

const (
	ReportSOC2     ReportType = "soc2"
	ReportISO42001 ReportType = "iso42001"
	ReportEUAIAct  ReportType = "eu-ai-act"
	ReportGDPR     ReportType = "gdpr"
)

// ReportRequest specifies what report to generate.
type ReportRequest struct {
	Type      ReportType `json:"type"`
	StartDate time.Time  `json:"start_date"`
	EndDate   time.Time  `json:"end_date"`
}

// Report is the generated compliance report.
type Report struct {
	ID           uuid.UUID        `json:"id"`
	Type         ReportType       `json:"type"`
	Title        string           `json:"title"`
	GeneratedAt  time.Time        `json:"generated_at"`
	Period       ReportPeriod     `json:"period"`
	Summary      ReportSummary    `json:"summary"`
	Controls     []ControlStatus  `json:"controls"`
	Findings     []Finding        `json:"findings"`
	HTMLContent  string           `json:"html_content,omitempty"`
}

// ReportPeriod defines the audit period.
type ReportPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ReportSummary provides aggregate statistics for the report.
type ReportSummary struct {
	TotalDecisions     int     `json:"total_decisions"`
	TotalReviews       int     `json:"total_reviews"`
	TotalEvidenceItems int     `json:"total_evidence_items"`
	AutomatedActions   map[string]int `json:"automated_actions"`
	HumanOverrides     int     `json:"human_overrides"`
	OverrideRate       float64 `json:"override_rate_pct"`
	AvgResponseTime    float64 `json:"avg_response_time_ms,omitempty"`
	PoliciesActive     int     `json:"policies_active"`
}

// ControlStatus reports on a single compliance control.
type ControlStatus struct {
	ControlID     string `json:"control_id"`
	Name          string `json:"name"`
	Status        string `json:"status"` // "effective", "partially_effective", "not_effective"
	EvidenceCount int    `json:"evidence_count"`
	Description   string `json:"description"`
}

// Finding represents a compliance finding or observation.
type Finding struct {
	Severity    string `json:"severity"` // "info", "low", "medium", "high", "critical"
	Category    string `json:"category"`
	Description string `json:"description"`
	Remediation string `json:"remediation,omitempty"`
}

// Generate creates a compliance report for the specified framework and period.
func (r *Reporter) Generate(ctx context.Context, req ReportRequest) (*Report, error) {
	report := &Report{
		ID:          uuid.New(),
		Type:        req.Type,
		GeneratedAt: time.Now().UTC(),
		Period:      ReportPeriod{Start: req.StartDate, End: req.EndDate},
	}

	// Gather statistics
	summary, err := r.gatherSummary(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to gather summary: %w", err)
	}
	report.Summary = *summary

	// Evaluate controls
	controls, err := r.evaluateControls(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate controls: %w", err)
	}
	report.Controls = controls

	// Generate findings
	report.Findings = r.generateFindings(summary, controls)

	// Set title and generate HTML based on report type
	switch req.Type {
	case ReportSOC2:
		report.Title = "SOC 2 Type II - AI Content Moderation Service"
		report.HTMLContent = r.renderSOC2HTML(report)
	case ReportISO42001:
		report.Title = "ISO/IEC 42001 - AI Management System Compliance"
		report.HTMLContent = r.renderISO42001HTML(report)
	case ReportEUAIAct:
		report.Title = "EU AI Act - High-Risk AI System Compliance"
		report.HTMLContent = r.renderEUAIActHTML(report)
	case ReportGDPR:
		report.Title = "GDPR - Data Protection Audit Report"
		report.HTMLContent = r.renderGDPRHTML(report)
	default:
		report.Title = fmt.Sprintf("Compliance Report - %s", req.Type)
		report.HTMLContent = r.renderGenericHTML(report)
	}

	r.logger.Info("compliance report generated",
		zap.String("type", string(req.Type)),
		zap.String("report_id", report.ID.String()),
		zap.Int("evidence_items", summary.TotalEvidenceItems),
	)

	return report, nil
}

func (r *Reporter) gatherSummary(ctx context.Context, start, end time.Time) (*ReportSummary, error) {
	summary := &ReportSummary{
		AutomatedActions: make(map[string]int),
	}

	// Total decisions
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM moderation_decisions WHERE created_at BETWEEN $1 AND $2`,
		start, end,
	).Scan(&summary.TotalDecisions)
	if err != nil {
		return nil, err
	}

	// Action breakdown
	rows, err := r.db.Query(ctx,
		`SELECT automated_action, COUNT(*) FROM moderation_decisions WHERE created_at BETWEEN $1 AND $2 GROUP BY automated_action`,
		start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			return nil, err
		}
		summary.AutomatedActions[action] = count
	}

	// Total reviews
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM review_actions WHERE created_at BETWEEN $1 AND $2`,
		start, end,
	).Scan(&summary.TotalReviews)
	if err != nil {
		return nil, err
	}

	// Human overrides (reviews that changed the automated decision)
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM review_actions ra
		 JOIN moderation_decisions md ON ra.decision_id = md.id
		 WHERE ra.created_at BETWEEN $1 AND $2
		 AND ra.action != md.automated_action`,
		start, end,
	).Scan(&summary.HumanOverrides)
	if err != nil {
		return nil, err
	}

	if summary.TotalReviews > 0 {
		summary.OverrideRate = float64(summary.HumanOverrides) / float64(summary.TotalReviews) * 100
	}

	// Evidence items
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM evidence_records WHERE created_at BETWEEN $1 AND $2`,
		start, end,
	).Scan(&summary.TotalEvidenceItems)
	if err != nil {
		return nil, err
	}

	// Active policies
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM policies WHERE status = 'published'`,
	).Scan(&summary.PoliciesActive)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *Reporter) evaluateControls(ctx context.Context, start, end time.Time) ([]ControlStatus, error) {
	// Query evidence counts by control_id
	rows, err := r.db.Query(ctx,
		`SELECT control_id, COUNT(*) FROM evidence_records WHERE created_at BETWEEN $1 AND $2 GROUP BY control_id`,
		start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evidenceCounts := make(map[string]int)
	for rows.Next() {
		var controlID string
		var count int
		if err := rows.Scan(&controlID, &count); err != nil {
			return nil, err
		}
		evidenceCounts[controlID] = count
	}

	// Define expected controls and evaluate
	controls := []struct {
		id   string
		name string
		desc string
	}{
		{"MOD-001", "Automated Text Classification", "AI model classifies submitted text and produces category scores"},
		{"MOD-003", "Moderation Request Pipeline", "End-to-end pipeline from submission to decision"},
		{"POL-001", "Threshold-Based Moderation Policy", "Policy engine evaluates scores against configurable thresholds"},
		{"GOV-001", "Role-Based Access Control", "RBAC enforcement with admin/moderator/viewer roles"},
		{"GOV-002", "Human-in-the-Loop Review", "Moderator review of escalated content"},
		{"AUD-001", "Immutable Evidence Storage", "Append-only evidence records for all decisions"},
		{"AUD-002", "Human Review Evidence", "Evidence records for human review actions"},
		{"SEC-002", "API Key Authentication", "SHA-256 hashed API key authentication"},
	}

	var result []ControlStatus
	for _, ctrl := range controls {
		count := evidenceCounts[ctrl.id]
		status := "effective"
		if count == 0 {
			status = "not_effective"
		} else if count < 10 {
			status = "partially_effective"
		}

		result = append(result, ControlStatus{
			ControlID:     ctrl.id,
			Name:          ctrl.name,
			Status:        status,
			EvidenceCount: count,
			Description:   ctrl.desc,
		})
	}

	return result, nil
}

func (r *Reporter) generateFindings(summary *ReportSummary, controls []ControlStatus) []Finding {
	var findings []Finding

	// Check for controls with no evidence
	for _, ctrl := range controls {
		if ctrl.Status == "not_effective" {
			findings = append(findings, Finding{
				Severity:    "high",
				Category:    "Control Gap",
				Description: fmt.Sprintf("Control %s (%s) has no evidence records in the audit period", ctrl.ControlID, ctrl.Name),
				Remediation: "Ensure the control is implemented and generating evidence records",
			})
		}
	}

	// Check human override rate
	if summary.OverrideRate > 30 {
		findings = append(findings, Finding{
			Severity:    "medium",
			Category:    "Model Performance",
			Description: fmt.Sprintf("Human override rate is %.1f%%, exceeding 30%% threshold", summary.OverrideRate),
			Remediation: "Review model accuracy and policy thresholds; consider retraining or adjusting thresholds",
		})
	}

	// Check if human reviews are happening for escalated content
	escalated := summary.AutomatedActions["escalate"]
	if escalated > 0 && summary.TotalReviews == 0 {
		findings = append(findings, Finding{
			Severity:    "high",
			Category:    "Governance Gap",
			Description: fmt.Sprintf("%d items were escalated but no human reviews were recorded", escalated),
			Remediation: "Ensure human reviewers are processing the review queue",
		})
	}

	if len(findings) == 0 {
		findings = append(findings, Finding{
			Severity:    "info",
			Category:    "Overall",
			Description: "No significant findings. All controls are operating effectively.",
		})
	}

	return findings
}

func (r *Reporter) renderSOC2HTML(report *Report) string {
	return r.renderHTML(report, "SOC 2 Type II", []string{
		"Trust Services Criteria: Security, Availability, Processing Integrity, Confidentiality",
		"Service Organization: Civitas AI Content Moderation Platform",
	})
}

func (r *Reporter) renderISO42001HTML(report *Report) string {
	return r.renderHTML(report, "ISO/IEC 42001:2023", []string{
		"AI Management System conformity assessment",
		"Clause 8: Operation - AI system lifecycle processes",
	})
}

func (r *Reporter) renderEUAIActHTML(report *Report) string {
	return r.renderHTML(report, "EU AI Act (Regulation 2024/1689)", []string{
		"Article 9: Risk Management System",
		"Article 13: Transparency and provision of information to deployers",
		"Article 14: Human oversight",
	})
}

func (r *Reporter) renderGDPRHTML(report *Report) string {
	return r.renderHTML(report, "GDPR (Regulation 2016/679)", []string{
		"Article 35: Data Protection Impact Assessment",
		"Article 22: Automated individual decision-making",
	})
}

func (r *Reporter) renderGenericHTML(report *Report) string {
	return r.renderHTML(report, string(report.Type), nil)
}

func (r *Reporter) renderHTML(report *Report, framework string, references []string) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>%s</title>
<style>
  body { font-family: 'Segoe UI', system-ui, sans-serif; max-width: 900px; margin: 0 auto; padding: 40px; color: #1a1a2e; }
  h1 { color: #16213e; border-bottom: 3px solid #0f3460; padding-bottom: 10px; }
  h2 { color: #0f3460; margin-top: 30px; }
  table { border-collapse: collapse; width: 100%%; margin: 20px 0; }
  th, td { border: 1px solid #ddd; padding: 10px; text-align: left; }
  th { background: #0f3460; color: white; }
  tr:nth-child(even) { background: #f8f9fa; }
  .effective { color: #28a745; font-weight: bold; }
  .partially_effective { color: #ffc107; font-weight: bold; }
  .not_effective { color: #dc3545; font-weight: bold; }
  .finding { border-left: 4px solid #ddd; padding: 10px; margin: 10px 0; }
  .finding.high { border-color: #dc3545; }
  .finding.medium { border-color: #ffc107; }
  .finding.low { border-color: #17a2b8; }
  .finding.info { border-color: #28a745; }
  .summary-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; margin: 20px 0; }
  .summary-card { background: #f8f9fa; border-radius: 8px; padding: 15px; text-align: center; }
  .summary-card .number { font-size: 2em; font-weight: bold; color: #0f3460; }
  .meta { color: #666; font-size: 0.9em; }
</style>
</head>
<body>
<h1>%s</h1>
<p class="meta">Framework: %s</p>
<p class="meta">Report ID: %s</p>
<p class="meta">Generated: %s</p>
<p class="meta">Audit Period: %s to %s</p>
`,
		report.Title, report.Title, framework,
		report.ID.String(),
		report.GeneratedAt.Format("2006-01-02 15:04:05 UTC"),
		report.Period.Start.Format("2006-01-02"),
		report.Period.End.Format("2006-01-02"),
	)

	if len(references) > 0 {
		html += "<h2>Framework References</h2><ul>"
		for _, ref := range references {
			html += fmt.Sprintf("<li>%s</li>", ref)
		}
		html += "</ul>"
	}

	// Summary section
	html += fmt.Sprintf(`
<h2>Executive Summary</h2>
<div class="summary-grid">
  <div class="summary-card"><div class="number">%d</div>Moderation Decisions</div>
  <div class="summary-card"><div class="number">%d</div>Human Reviews</div>
  <div class="summary-card"><div class="number">%d</div>Evidence Records</div>
  <div class="summary-card"><div class="number">%d</div>Active Policies</div>
  <div class="summary-card"><div class="number">%d</div>Human Overrides</div>
  <div class="summary-card"><div class="number">%.1f%%</div>Override Rate</div>
</div>
`,
		report.Summary.TotalDecisions,
		report.Summary.TotalReviews,
		report.Summary.TotalEvidenceItems,
		report.Summary.PoliciesActive,
		report.Summary.HumanOverrides,
		report.Summary.OverrideRate,
	)

	// Action breakdown
	html += "<h2>Action Breakdown</h2><table><tr><th>Action</th><th>Count</th></tr>"
	for action, count := range report.Summary.AutomatedActions {
		html += fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", action, count)
	}
	html += "</table>"

	// Controls table
	html += "<h2>Control Assessment</h2><table><tr><th>Control ID</th><th>Name</th><th>Status</th><th>Evidence Count</th></tr>"
	for _, ctrl := range report.Controls {
		html += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td class="%s">%s</td><td>%d</td></tr>`,
			ctrl.ControlID, ctrl.Name, ctrl.Status, ctrl.Status, ctrl.EvidenceCount)
	}
	html += "</table>"

	// Findings
	html += "<h2>Findings</h2>"
	for _, f := range report.Findings {
		html += fmt.Sprintf(`<div class="finding %s"><strong>[%s] %s:</strong> %s`,
			f.Severity, f.Severity, f.Category, f.Description)
		if f.Remediation != "" {
			html += fmt.Sprintf(`<br><em>Remediation:</em> %s`, f.Remediation)
		}
		html += "</div>"
	}

	html += `
<hr>
<p class="meta">This report was automatically generated by the Civitas AI Compliance Engine.
Evidence records are immutable and cryptographically verifiable.</p>
</body>
</html>`

	return html
}
