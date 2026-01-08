package database

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

var pool *pgxpool.Pool

// Init db connection pool
func Connect() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal().Msg("DATABASE_URL environment variable is required")
	}

	// Connection pool configs
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return err
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return err
	}
	log.Info().Msg("Connected to PostgreSQL database")
	return nil
}

// Shuts down the connection pool
func Close() {
	if pool != nil {
		pool.Close()
		log.Info().Msg("Database connection closed")
	}
}

// Return connection pool for queries
func getPool() *pgxpool.Pool {
	return pool
}
