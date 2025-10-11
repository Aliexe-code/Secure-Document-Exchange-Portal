package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load database config : %w", err)
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("Database Connected :)")
	return pool, nil
}
