package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const keyPrefix = "tm_live_"

// KeyInfo represents non-sensitive API key metadata.
type KeyInfo struct {
	UserID     uuid.UUID  `json:"user_id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RateLimitRPM int     `json:"rate_limit_rpm"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Manager handles API key lifecycle operations.
// Control: SEC-002 (API Key and OAuth Authentication)
type Manager struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewManager creates a new API key manager.
func NewManager(pool *pgxpool.Pool, logger *zap.Logger) *Manager {
	return &Manager{pool: pool, logger: logger}
}

// GenerateKey creates a new API key for a user.
// Returns the plaintext key (only shown once) — the database stores only the hash.
func (m *Manager) GenerateKey(ctx context.Context, userID uuid.UUID, name string) (string, error) {
	// Generate 32 random bytes for the key
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	plaintext := keyPrefix + hex.EncodeToString(randomBytes)
	prefix := plaintext[:len(keyPrefix)+8]
	hash := hashKey(plaintext)

	_, err := m.pool.Exec(ctx,
		`UPDATE users SET api_key_hash = $1, api_key_name = $2, api_key_prefix = $3 WHERE id = $4`,
		hash, name, prefix, userID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store API key: %w", err)
	}

	m.logger.Info("API key generated", zap.String("user_id", userID.String()), zap.String("prefix", prefix))
	return plaintext, nil
}

// RevokeKey revokes the API key for a user.
func (m *Manager) RevokeKey(ctx context.Context, userID uuid.UUID) error {
	result, err := m.pool.Exec(ctx,
		`UPDATE users SET api_key_hash = NULL, api_key_name = NULL, api_key_prefix = NULL WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	m.logger.Info("API key revoked", zap.String("user_id", userID.String()))
	return nil
}

// RotateKey revokes the old key and generates a new one atomically.
func (m *Manager) RotateKey(ctx context.Context, userID uuid.UUID, name string) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	plaintext := keyPrefix + hex.EncodeToString(randomBytes)
	prefix := plaintext[:len(keyPrefix)+8]
	hash := hashKey(plaintext)

	result, err := m.pool.Exec(ctx,
		`UPDATE users SET api_key_hash = $1, api_key_name = $2, api_key_prefix = $3 WHERE id = $4`,
		hash, name, prefix, userID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to rotate API key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return "", fmt.Errorf("user not found")
	}

	m.logger.Info("API key rotated", zap.String("user_id", userID.String()), zap.String("prefix", prefix))
	return plaintext, nil
}

// ListKeys returns non-sensitive key metadata for all users with keys.
func (m *Manager) ListKeys(ctx context.Context) ([]KeyInfo, error) {
	rows, err := m.pool.Query(ctx,
		`SELECT id, COALESCE(api_key_name, ''), COALESCE(api_key_prefix, ''), api_key_last_used_at, COALESCE(rate_limit_rpm, 60), created_at
		 FROM users WHERE api_key_hash IS NOT NULL
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	defer rows.Close()

	var keys []KeyInfo
	for rows.Next() {
		var k KeyInfo
		if err := rows.Scan(&k.UserID, &k.Name, &k.Prefix, &k.LastUsedAt, &k.RateLimitRPM, &k.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// GetKeyRateLimit returns the per-key RPM for a given API key hash.
func (m *Manager) GetKeyRateLimit(ctx context.Context, apiKeyHash string) (int, error) {
	var rpm int
	err := m.pool.QueryRow(ctx,
		`SELECT COALESCE(rate_limit_rpm, 60) FROM users WHERE api_key_hash = $1`,
		apiKeyHash,
	).Scan(&rpm)
	if err != nil {
		return 0, fmt.Errorf("key not found: %w", err)
	}
	return rpm, nil
}

// UpdateLastUsed records when a key was last used.
func (m *Manager) UpdateLastUsed(ctx context.Context, apiKeyHash string) {
	_, err := m.pool.Exec(ctx,
		`UPDATE users SET api_key_last_used_at = NOW() WHERE api_key_hash = $1`,
		apiKeyHash,
	)
	if err != nil {
		m.logger.Warn("failed to update key last used", zap.Error(err))
	}
}

func hashKey(plaintext string) string {
	h := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(h[:])
}
