package pgretry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	delays      = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	maxAttempts = 3
)

func Retry[T any](ctx context.Context, op string, fn func() (T, error)) (T, error) {
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
