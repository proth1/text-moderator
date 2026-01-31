package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/proth1/text-moderator/internal/models"
	"go.uber.org/zap"
)

// Dispatcher handles webhook event dispatch and delivery.
// Control: INT-001 (Webhook event notification system)
type Dispatcher struct {
	db         *pgxpool.Pool
	httpClient *http.Client
	logger     *zap.Logger
}

// NewDispatcher creates a new webhook dispatcher.
func NewDispatcher(db *pgxpool.Pool, logger *zap.Logger) *Dispatcher {
	return &Dispatcher{
		db: db,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Dispatch sends an event to all active subscriptions matching the event type.
// Delivery is asynchronous - failures are recorded and retried.
func (d *Dispatcher) Dispatch(ctx context.Context, eventType models.WebhookEventType, data interface{}) error {
	// Build payload
	payload := models.WebhookPayload{
		ID:        uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Find active subscriptions matching this event type
	subs, err := d.findSubscriptions(ctx, string(eventType))
	if err != nil {
		return fmt.Errorf("failed to find subscriptions: %w", err)
	}

	if len(subs) == 0 {
		return nil
	}

	d.logger.Info("dispatching webhook event",
		zap.String("event_type", string(eventType)),
		zap.Int("subscriber_count", len(subs)),
	)

	// Deliver to each subscriber asynchronously
	for _, sub := range subs {
		go d.deliver(context.Background(), sub, string(eventType), payloadJSON)
	}

	return nil
}

// deliver attempts to send the webhook payload to a single subscriber.
func (d *Dispatcher) deliver(ctx context.Context, sub models.WebhookSubscription, eventType string, payloadJSON []byte) {
	deliveryID := uuid.New()
	maxAttempts := 5

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Compute HMAC-SHA256 signature
		signature := computeHMAC(payloadJSON, sub.Secret)

		req, err := http.NewRequestWithContext(ctx, "POST", sub.URL, bytes.NewBuffer(payloadJSON))
		if err != nil {
			d.logger.Error("failed to create webhook request",
				zap.String("subscription_id", sub.ID.String()),
				zap.Error(err),
			)
			break
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-ID", deliveryID.String())
		req.Header.Set("X-Webhook-Event", eventType)
		req.Header.Set("X-Webhook-Signature", "sha256="+signature)
		req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

		resp, err := d.httpClient.Do(req)
		var statusCode int
		var respBody string

		if err != nil {
			d.logger.Warn("webhook delivery failed",
				zap.String("subscription_id", sub.ID.String()),
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
		} else {
			statusCode = resp.StatusCode
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			respBody = string(body)

			if statusCode >= 200 && statusCode < 300 {
				// Successful delivery
				d.recordDelivery(ctx, deliveryID, sub.ID, eventType, string(payloadJSON),
					&statusCode, &respBody, attempt, maxAttempts, true)
				d.logger.Info("webhook delivered",
					zap.String("subscription_id", sub.ID.String()),
					zap.String("event_type", eventType),
					zap.Int("status", statusCode),
				)
				return
			}
		}

		// Record failed attempt
		if attempt == maxAttempts {
			d.recordDelivery(ctx, deliveryID, sub.ID, eventType, string(payloadJSON),
				intPtr(statusCode), strPtr(respBody), attempt, maxAttempts, false)
			d.logger.Error("webhook delivery exhausted all attempts",
				zap.String("subscription_id", sub.ID.String()),
				zap.String("event_type", eventType),
			)
			return
		}

		// Exponential backoff: 1s, 4s, 16s, 64s
		backoff := time.Duration(1<<uint(attempt-1)) * time.Second
		if backoff > 64*time.Second {
			backoff = 64 * time.Second
		}
		time.Sleep(backoff)
	}
}

func (d *Dispatcher) findSubscriptions(ctx context.Context, eventType string) ([]models.WebhookSubscription, error) {
	query := `
		SELECT id, url, secret, event_types, active, description, created_by, created_at, updated_at
		FROM webhook_subscriptions
		WHERE active = true AND $1 = ANY(event_types)
	`
	rows, err := d.db.Query(ctx, query, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.WebhookSubscription
	for rows.Next() {
		var sub models.WebhookSubscription
		if err := rows.Scan(
			&sub.ID, &sub.URL, &sub.Secret, &sub.EventTypes,
			&sub.Active, &sub.Description, &sub.CreatedBy,
			&sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

func (d *Dispatcher) recordDelivery(ctx context.Context, deliveryID, subID uuid.UUID, eventType, payload string, status *int, respBody *string, attempt, maxAttempts int, success bool) {
	query := `
		INSERT INTO webhook_deliveries (
			id, subscription_id, event_type, payload,
			response_status, response_body, attempt, max_attempts,
			delivered_at, failed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var deliveredAt, failedAt *time.Time
	now := time.Now().UTC()
	if success {
		deliveredAt = &now
	} else {
		failedAt = &now
	}

	_, err := d.db.Exec(ctx, query,
		deliveryID, subID, eventType, payload,
		status, respBody, attempt, maxAttempts,
		deliveredAt, failedAt,
	)
	if err != nil {
		d.logger.Error("failed to record webhook delivery",
			zap.String("delivery_id", deliveryID.String()),
			zap.Error(err),
		)
	}
}

// CreateSubscription creates a new webhook subscription.
func (d *Dispatcher) CreateSubscription(ctx context.Context, req *models.CreateWebhookRequest, createdBy *uuid.UUID) (*models.WebhookSubscription, error) {
	sub := &models.WebhookSubscription{
		ID:         uuid.New(),
		URL:        req.URL,
		Secret:     generateSecret(),
		EventTypes: req.EventTypes,
		Active:     true,
		CreatedBy:  createdBy,
	}
	if req.Description != "" {
		sub.Description = &req.Description
	}

	query := `
		INSERT INTO webhook_subscriptions (id, url, secret, event_types, active, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	err := d.db.QueryRow(ctx, query,
		sub.ID, sub.URL, sub.Secret, sub.EventTypes,
		sub.Active, sub.Description, sub.CreatedBy,
	).Scan(&sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create webhook subscription: %w", err)
	}

	return sub, nil
}

// ListSubscriptions returns all webhook subscriptions.
func (d *Dispatcher) ListSubscriptions(ctx context.Context) ([]models.WebhookSubscription, error) {
	query := `
		SELECT id, url, secret, event_types, active, description, created_by, created_at, updated_at
		FROM webhook_subscriptions
		ORDER BY created_at DESC
	`
	rows, err := d.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.WebhookSubscription
	for rows.Next() {
		var sub models.WebhookSubscription
		if err := rows.Scan(
			&sub.ID, &sub.URL, &sub.Secret, &sub.EventTypes,
			&sub.Active, &sub.Description, &sub.CreatedBy,
			&sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

// DeleteSubscription deactivates a webhook subscription.
func (d *Dispatcher) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE webhook_subscriptions SET active = false WHERE id = $1`
	_, err := d.db.Exec(ctx, query, id)
	return err
}

// computeHMAC generates an HMAC-SHA256 signature for the payload.
func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// generateSecret creates a random secret for HMAC signing.
func generateSecret() string {
	return uuid.New().String() + "-" + uuid.New().String()
}

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }
