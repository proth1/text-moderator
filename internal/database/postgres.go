package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PostgresDB wraps pgxpool for database operations
// Control: AUD-001 (Database connection management)
type PostgresDB struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// Config holds database configuration
type Config struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(ctx context.Context, cfg Config, logger *zap.Logger) (*PostgresDB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set pool configuration
	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	} else {
		poolConfig.MaxConns = 25 // Default max connections
	}

	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	} else {
		poolConfig.MinConns = 5 // Default min connections
	}

	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	} else {
		poolConfig.MaxConnLifetime = time.Hour
	}

	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	} else {
		poolConfig.MaxConnIdleTime = 30 * time.Minute
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established",
		zap.Int32("max_conns", poolConfig.MaxConns),
		zap.Int32("min_conns", poolConfig.MinConns),
	)

	return &PostgresDB{
		Pool:   pool,
		logger: logger,
	}, nil
}

// Close closes the database connection pool
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("database connection closed")
	}
}

// Health checks the database health
func (db *PostgresDB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Stats returns connection pool statistics
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}
