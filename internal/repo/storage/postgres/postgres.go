package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgStorage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dataBaseDSN string) (*PgStorage, error) {

	pool, err := pgxpool.New(ctx, dataBaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool, %w", err)
	}

	return &PgStorage{db: pool}, nil
}

func (p *PgStorage) Ping(ctx context.Context) error {
	return p.db.Ping(ctx)
}
