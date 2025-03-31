package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/zetcan333/metrics-collector/internal/models"
)

// Структура агента
type Agent struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
	Metrics        map[string]models.Metrics
	PollCount      int64
	client         http.Client
	sync.RWMutex
}

type MetricsSnapshot struct {
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64
	RandomValue   float64
	PollCount     int64
}

// Конструктор агента
func NewAgent(serverURL string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		ServerURL:      serverURL,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Metrics:        make(map[string]models.Metrics),
		client: http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (a *Agent) Start(ctx context.Context) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		ticker := time.NewTicker(a.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.CollectMetrics()
			case <-ctx.Done():
				fmt.Println("Metrics collection stopped")
				return
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(a.ReportInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := a.SendMetrics(); err != nil {
					fmt.Printf("Error sending metrics: %v\n", err)
				}
			case <-ctx.Done():
				fmt.Println("Metrics reporting stopped")
				return
			}
		}
	}()
	select {
	case <-stop:
		fmt.Println("Received termination signal, shutting down...")
		cancel()
	case <-ctx.Done():
	}
	time.Sleep(1 * time.Second)

}

// Отправляем метрики на сервер
func (a *Agent) SendMetrics() error {
	a.RLock()
	defer a.RUnlock()

	baseURL, err := url.Parse(a.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}
	baseURL.Path = path.Join(baseURL.Path, "update")

	for name, value := range a.Metrics {
		body, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to encode metric %v", err)
		}

		compressedBody, err := compressData(body)
		if err != nil {
			return fmt.Errorf("failed to compress data: %v", err)
		}

		req, err := http.NewRequest("POST", baseURL.String(), bytes.NewBuffer(compressedBody))
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}

		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		resp, err := a.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %v", err)
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d for metric %s", resp.StatusCode, name)
		}
	}
	return nil
}

// Сбор метрик из runtime
func (a *Agent) CollectMetrics() {
	a.Lock()
	defer a.Unlock()
	a.PollCount++
	snapshot := MetricsSnapshot{}
	snapshot.collectFlat(a.PollCount)

	v := reflect.ValueOf(snapshot)
	t := v.Type()

	for i := range v.NumField() {
		field := v.Field(i)
		name := t.Field(i).Name

		if name == "PollCount" {
			delta := field.Int()
			a.Metrics[name] = models.Metrics{ID: name, MType: models.Counter, Delta: &delta}
			continue
		}

		if field.Kind() == reflect.Float64 {
			value := field.Float()
			a.Metrics[name] = models.Metrics{ID: name, MType: models.Gauge, Value: &value}
		}
	}
}

func (m *MetricsSnapshot) collectFlat(pollCount int64) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	m.Alloc = float64(rtm.Alloc)
	m.BuckHashSys = float64(rtm.BuckHashSys)
	m.Frees = float64(rtm.Frees)
	m.GCCPUFraction = float64(rtm.GCCPUFraction)
	m.GCSys = float64(rtm.GCSys)
	m.HeapAlloc = float64(rtm.HeapAlloc)
	m.HeapIdle = float64(rtm.HeapIdle)
	m.HeapInuse = float64(rtm.HeapInuse)
	m.HeapObjects = float64(rtm.HeapObjects)
	m.HeapReleased = float64(rtm.HeapReleased)
	m.HeapSys = float64(rtm.HeapSys)
	m.LastGC = float64(rtm.LastGC)
	m.Lookups = float64(rtm.Lookups)
	m.MCacheInuse = float64(rtm.MCacheInuse)
	m.MCacheSys = float64(rtm.MCacheSys)
	m.MSpanInuse = float64(rtm.MSpanInuse)
	m.MSpanSys = float64(rtm.MSpanSys)
	m.Mallocs = float64(rtm.Mallocs)
	m.NextGC = float64(rtm.NextGC)
	m.NumForcedGC = float64(rtm.NumForcedGC)
	m.NumGC = float64(rtm.NumGC)
	m.OtherSys = float64(rtm.OtherSys)
	m.PauseTotalNs = float64(rtm.PauseTotalNs)
	m.StackInuse = float64(rtm.StackInuse)
	m.StackSys = float64(rtm.StackSys)
	m.Sys = float64(rtm.Sys)

	m.RandomValue = rand.Float64()
	m.PollCount = pollCount
}

func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
