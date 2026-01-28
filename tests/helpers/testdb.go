package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB provides a test database setup with migrations and seed data
type TestDB struct {
	Pool *pgxpool.Pool
	DSN  string
}

// SetupTestDB creates a test database connection, runs migrations, and loads seed data
func SetupTestDB(t *testing.T) (*TestDB, func()) {
	t.Helper()

	// Get database URL from environment or use default
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/civitas_test?sslmode=disable"
	}

	// Create connection pool
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("Failed to parse database config: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("Failed to ping database: %v", err)
	}

	testDB := &TestDB{
		Pool: pool,
		DSN:  dsn,
	}

	// Run migrations (assumes migrations are handled elsewhere)
	// In production, this would call migrate.Up() or similar

	// Load seed data
	if err := testDB.LoadSeedData(t); err != nil {
		pool.Close()
		t.Fatalf("Failed to load seed data: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		testDB.Cleanup(t)
		pool.Close()
	}

	return testDB, cleanup
}

// LoadSeedData loads test fixtures from seed-data.sql
func (db *TestDB) LoadSeedData(t *testing.T) error {
	t.Helper()

	// Find the seed data file
	seedFile := filepath.Join("tests", "fixtures", "seed-data.sql")
	if _, err := os.Stat(seedFile); os.IsNotExist(err) {
		// Try alternative path
		seedFile = filepath.Join("..", "fixtures", "seed-data.sql")
	}

	data, err := os.ReadFile(seedFile)
	if err != nil {
		return fmt.Errorf("failed to read seed data: %w", err)
	}

	ctx := context.Background()
	if _, err := db.Pool.Exec(ctx, string(data)); err != nil {
		return fmt.Errorf("failed to execute seed data: %w", err)
	}

	return nil
}

// Cleanup removes all test data from the database
func (db *TestDB) Cleanup(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	// Delete in reverse order of foreign key dependencies
	tables := []string{
		"evidence_records",
		"review_actions",
		"moderation_decisions",
		"text_submissions",
		"policies",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			t.Logf("Warning: Failed to clean table %s: %v", table, err)
		}
	}
}

// TruncateAll truncates all tables (faster than DELETE but requires TRUNCATE permissions)
func (db *TestDB) TruncateAll(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	tables := []string{
		"evidence_records",
		"review_actions",
		"moderation_decisions",
		"text_submissions",
		"policies",
		"users",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// GetTestUser returns a test user by email
func (db *TestDB) GetTestUser(t *testing.T, email string) (userID string, apiKey string) {
	t.Helper()

	ctx := context.Background()
	query := "SELECT id, api_key FROM users WHERE email = $1"

	var key *string
	err := db.Pool.QueryRow(ctx, query, email).Scan(&userID, &key)
	if err != nil {
		t.Fatalf("Failed to get test user %s: %v", email, err)
	}

	if key != nil {
		apiKey = *key
	}

	return userID, apiKey
}

// CreateTestSubmission creates a test text submission
func (db *TestDB) CreateTestSubmission(t *testing.T, content string, contentHash string) string {
	t.Helper()

	ctx := context.Background()
	query := `
		INSERT INTO text_submissions (content_hash, content_encrypted, context_metadata, source)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var submissionID string
	err := db.Pool.QueryRow(ctx, query, contentHash, content, `{"test": true}`, "test").Scan(&submissionID)
	if err != nil {
		t.Fatalf("Failed to create test submission: %v", err)
	}

	return submissionID
}

// WaitForCondition polls a condition until it's true or timeout
func WaitForCondition(t *testing.T, timeout time.Duration, interval time.Duration, condition func() bool) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}
