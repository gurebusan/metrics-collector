package postgres

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zetcan333/metrics-collector/internal/models"
	"github.com/zetcan333/metrics-collector/pkg/myerrors"
)

type PgStorage struct {
	db *pgxpool.Pool
}

var (
	delays      = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	maxAttempts = 3
)

func NewStorage(ctx context.Context, dataBaseDSN string) (*PgStorage, error) {
	const op = "internal.repo.storage.postgres.NewStorage"

	pool, err := retry(op, ctx, maxAttempts, delays, func() (*pgxpool.Pool, error) {
		pool, err := pgxpool.New(ctx, dataBaseDSN)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close() // важно
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return pool, nil
	})

	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS metrics (
        ID TEXT PRIMARY KEY,
		type TEXT NOT NULL,
        value DOUBLE PRECISION NOT NULL,
        delta INT8 NOT NULL
    );
	`

	_, err = pool.Exec(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PgStorage{db: pool}, nil
}

func (p *PgStorage) Ping(ctx context.Context) error {
	return p.db.Ping(ctx)
}

func (p *PgStorage) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	const op = "internal.repo.storage.postgres.UpdateMetric"

	_, err := retry(op, ctx, maxAttempts, delays, func() (struct{}, error) {

		var err error

		switch metric.MType {

		case models.Gauge:
			_, err = p.db.Exec(ctx, `
			INSERT INTO metrics (id, type, value, delta) 
			VALUES ($1, $2, $3, 0) 
			ON CONFLICT (id) DO UPDATE 
			SET value = $3, type = $2
		`, metric.ID, metric.MType, *metric.Value)

		case models.Counter:
			_, err = p.db.Exec(ctx, `
			INSERT INTO metrics (id, type, value, delta)
			VALUES ($1, $2, 0, $3)
			ON CONFLICT (id) DO UPDATE
			SET delta = metrics.delta + $3, type = $2
		`, metric.ID, metric.MType, *metric.Delta)
		}
		if err != nil {
			return struct{}{}, fmt.Errorf("%s: %w", op, err)
		}
		return struct{}{}, nil
	})
	return err
}
func (p *PgStorage) GetMetric(ctx context.Context, id string) (models.Metrics, error) {
	const op = "internal.repo.storage.postgres.GetMetric"
	var (
		metric     models.Metrics
		metricType string
		value      float64
		delta      int64
	)
	return retry(op, ctx, maxAttempts, delays, func() (models.Metrics, error) {
		err := p.db.QueryRow(ctx, `SELECT id, type, value, delta FROM metrics WHERE id = $1`, id).
			Scan(&metric.ID, &metricType, &value, &delta)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return models.Metrics{}, myerrors.ErrMetricNotFound
			}
			return models.Metrics{}, fmt.Errorf("%s: %w", op, err)
		}
		metric.MType = metricType

		if metricType == models.Gauge {
			metric.Value = &value
		} else if metricType == models.Counter {
			metric.Delta = &delta
		}
		return metric, nil
	})

}

func (p *PgStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	const op = "internal.repo.storage.postgres.GetAllGauges"

	return retry(op, ctx, maxAttempts, delays, func() (map[string]float64, error) {
		rows, err := p.db.Query(ctx, `
		SELECT id, value FROM metrics WHERE type = $1
	`, models.Gauge)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		defer rows.Close()

		gauges := make(map[string]float64)
		for rows.Next() {
			var id string
			var value float64
			if err := rows.Scan(&id, &value); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			gauges[id] = value
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return gauges, nil
	})
}

func (p *PgStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	const op = "internal.repo.storage.postgres.GetAllCounters"

	return retry(op, ctx, maxAttempts, delays, func() (map[string]int64, error) {
		rows, err := p.db.Query(ctx, `
	SELECT id, delta FROM metrics WHERE type = $1
	`, models.Counter)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		defer rows.Close()
		counters := make(map[string]int64)
		for rows.Next() {
			var id string
			var delta int64
			if err := rows.Scan(&id, &delta); err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			counters[id] = delta
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return counters, nil
	})
}

func (p *PgStorage) UpdateMetricsWithBatch(ctx context.Context, metrics []models.Metrics) error {
	const op = "internal.repo.storage.postgres.UpdateMetricsWithBatch"

	_, err := retry(op, ctx, maxAttempts, delays, func() (struct{}, error) {

		tx, err := p.db.Begin(ctx)
		if err != nil {
			return struct{}{}, fmt.Errorf("%s: %w", op, err)
		}
		defer tx.Rollback(ctx)
		for _, metric := range metrics {
			var err error
			switch metric.MType {
			case models.Gauge:
				_, err = tx.Exec(ctx, `
				INSERT INTO metrics (id, type, value, delta) 
				VALUES ($1, $2, $3, 0) 
				ON CONFLICT (id) DO UPDATE 
				SET value = $3, type = $2
			`, metric.ID, metric.MType, *metric.Value)
			case models.Counter:
				_, err = tx.Exec(ctx, `
				INSERT INTO metrics (id, type, value, delta)
				VALUES ($1, $2, 0, $3)
				ON CONFLICT (id) DO UPDATE
				SET delta = metrics.delta + $3, type = $2
			`, metric.ID, metric.MType, *metric.Delta)
			}
			if err != nil {
				tx.Rollback(ctx)
				return struct{}{}, fmt.Errorf("%s: %w", op, err)
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return struct{}{}, fmt.Errorf("%s: %w", op, err)
		}
		return struct{}{}, nil
	})
	return err
}

// mock SaveBkpToFile and LoadBkpFromFile
func (p *PgStorage) SaveBkpToFile(path string) error {
	return nil
}
func (p *PgStorage) LoadBkpFromFile(path string) error {
	return nil
}

func retry[T any](op string, ctx context.Context, maxAttempts int, delays []time.Duration, fn func() (T, error)) (T, error) {
	var zero T

	for attempt := 0; attempt < maxAttempts; attempt++ {
		result, err := fn()

		switch {

		case err == nil:
			return result, nil

		case !isRetriableError(err):
			return zero, fmt.Errorf("not retriable error: %w", err)
		}

		fmt.Printf("op: %s, retrying to exec, attempt: %d\n", op, attempt+1)
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delays[attempt]):
		}
	}
	return zero, fmt.Errorf("max attempts reached")
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection:
			return true
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	if strings.Contains(err.Error(), "dial tcp") {
		return true
	}

	return false
}
