package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/zetcan333/metrics-collector/internal/flags"
	"github.com/zetcan333/metrics-collector/internal/models"
)

var (
	delays      = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	maxAttempts = 3
)

// Структура агента
type Agent struct {
	flags     *flags.AgentFlags
	Metrics   map[string]models.Metrics
	PollCount int64
	client    http.Client
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
func NewAgent(flags *flags.AgentFlags) *Agent {
	return &Agent{
		flags:   flags,
		Metrics: make(map[string]models.Metrics),
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
		ticker := time.NewTicker(a.flags.PollInterval)
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
		ticker := time.NewTicker(a.flags.ReportInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := a.SendMetricsBatch(); err != nil {
					fmt.Printf("Error sending metrics: %v\n", err)
				} else {
					fmt.Println("Metrics sent successfully")
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

}

// DEPRICATED
func (a *Agent) SendMetrics() error {
	a.RLock()
	defer a.RUnlock()

	baseURL, err := url.Parse(a.flags.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	updateURL := baseURL.JoinPath("update")

	for name, value := range a.Metrics {
		body, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to encode metric %v", err)
		}

		compressedBody, err := compressData(body)
		if err != nil {
			return fmt.Errorf("failed to compress data: %v", err)
		}

		req, err := http.NewRequest("POST", updateURL.String(), bytes.NewBuffer(compressedBody))
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

// ACTUAL FOR CURRENT API
func (a *Agent) SendMetricsBatch() error {
	a.RLock()
	defer a.RUnlock()

	if len(a.Metrics) == 0 {
		return nil
	}

	baseURL, err := url.Parse(a.flags.ServerURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	updateURL := baseURL.JoinPath("updates/")

	metrics := make([]models.Metrics, 0, len(a.Metrics))
	for _, metric := range a.Metrics {
		metrics = append(metrics, metric)
	}

	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to encode metrics: %v", err)
	}

	compressedBody, err := compressData(body)
	if err != nil {
		return fmt.Errorf("failed to compress data: %v", err)
	}

	req, err := http.NewRequest("POST", updateURL.String(), bytes.NewBuffer(compressedBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Encoding", "gzip")
	if a.flags.Key != "" {
		hash := createHash(body, a.flags.Key)
		req.Header.Set("HashSHA256", hash)
	}
	req.Header.Set("Content-Type", "application/json")

	return retry(maxAttempts, delays, func() error {

		resp, err := a.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		return nil
	})
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
			if existing, ok := a.Metrics[name]; ok && existing.Delta != nil {
				newDelta := *existing.Delta + 1
				a.Metrics[name] = models.Metrics{ID: name, MType: models.Counter, Delta: &newDelta}
			} else {
				initial := int64(1)
				a.Metrics[name] = models.Metrics{ID: name, MType: models.Counter, Delta: &initial}
			}
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
	m.TotalAlloc = float64(rtm.TotalAlloc)

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

func createHash(data []byte, key string) string {
	fmt.Println("agent", key)
	fmt.Println(data)
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func retry(maxAttempts int, delays []time.Duration, fn func() error) error {

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := fn()

		switch {
		case err == nil:
			return nil
		case !isRetriableError(err):
			return err
		}
		fmt.Printf("retrying to send, attempt: %d\n", attempt+1)
		<-time.After(delays[attempt])
	}
	return fmt.Errorf("max attempts reached")
}

func isRetriableError(err error) bool {
	var netErr net.Error

	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if strings.Contains(err.Error(), "dial tcp") {
		return true
	}

	return false
}
