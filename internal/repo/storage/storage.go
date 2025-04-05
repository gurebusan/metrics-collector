package storage

import (
	"context"

	"github.com/zetcan333/metrics-collector/internal/models"
)

type Storage interface {
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	GetMetric(ctx context.Context, id string) (models.Metrics, error)
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
	UpdateMetricsWithBatch(ctx context.Context, metrics []models.Metrics) error
	SaveBkpToFile(path string) error
	LoadBkpFromFile(path string) error
	Ping(ctx context.Context) error
}
