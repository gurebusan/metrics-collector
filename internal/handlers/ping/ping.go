package ping

import (
	"context"
	"net/http"
	"time"
)

type PgPing interface {
	Ping(ctx context.Context) error
}

type PingHandler struct {
	pg PgPing
}

func New(p PgPing) *PingHandler {
	return &PingHandler{pg: p}
}

func (p *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := p.pg.Ping(ctx); err != nil {
		http.Error(w, "database connection failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
