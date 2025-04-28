package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
)

func Open(ctx context.Context, cfg config.Config) (DB, error) {
	pool, err := pgxpool.Connect(ctx, generateDsn(cfg))
	if err != nil {
		return nil, err
	}

	return newPostgresDatabase(pool), nil
}

func generateDsn(config config.Config) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost,
		config.DBPort,
		config.DBUser,
		config.DBPass,
		config.DBName,
	)
}
