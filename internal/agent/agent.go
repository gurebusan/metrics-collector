package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
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

// Конструктор агента
func NewAgent(serverURL string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		ServerURL:      serverURL,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Metrics:        make(map[string]models.Metrics),
	}
}

// Запуск агента
func (a *Agent) Start(ctx context.Context) {
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
				err := a.SendMetrics()
				if err != nil {
					fmt.Println(err)
				}
			case <-ctx.Done():
				fmt.Println("Metrics reporting stopped")
				return
			}
		}
	}()
}

// Отправляем метрики на сервер
func (a *Agent) SendMetrics() error {
	a.RLock()
	defer a.RUnlock()

	for name, value := range a.Metrics {
		body, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to encode metric %v", err)
		}

		compressedBody, err := compressData(body)
		if err != nil {
			return fmt.Errorf("failed to compress data: %v", err)
		}

		destination := fmt.Sprintf("%s/update/", a.ServerURL)

		req, err := http.NewRequest("POST", destination, bytes.NewBuffer(compressedBody))
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
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	a.Lock()
	defer a.Unlock()

	alloc := float64(rtm.Alloc)
	a.Metrics["Alloc"] = models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &alloc}

	buckHashSys := float64(rtm.Frees)
	a.Metrics["BuckHashSys"] = models.Metrics{ID: "BuckHashSys", MType: models.Gauge, Value: &buckHashSys}

	frees := float64(rtm.Frees)
	a.Metrics["Frees"] = models.Metrics{ID: "Frees", MType: models.Gauge, Value: &frees}

	gCCPUFraction := rtm.GCCPUFraction
	a.Metrics["GCCPUFraction"] = models.Metrics{ID: "GCCPUFraction", MType: models.Gauge, Value: &gCCPUFraction}

	gCSys := float64(rtm.GCSys)
	a.Metrics["GCSys"] = models.Metrics{ID: "GCSys", MType: models.Gauge, Value: &gCSys}

	heapAlloc := float64(rtm.HeapAlloc)
	a.Metrics["HeapAlloc"] = models.Metrics{ID: "HeapAlloc", MType: models.Gauge, Value: &heapAlloc}

	heapIdle := float64(rtm.HeapIdle)
	a.Metrics["HeapIdle"] = models.Metrics{ID: "HeapIdle", MType: models.Gauge, Value: &heapIdle}

	heapInuse := float64(rtm.HeapInuse)
	a.Metrics["HeapInuse"] = models.Metrics{ID: "HeapInuse", MType: models.Gauge, Value: &heapInuse}

	heapObjects := float64(rtm.HeapObjects)
	a.Metrics["HeapObjects"] = models.Metrics{ID: "HeapObjects", MType: models.Gauge, Value: &heapObjects}

	heapReleased := float64(rtm.HeapReleased)
	a.Metrics["HeapReleased"] = models.Metrics{ID: "HeapReleased", MType: models.Gauge, Value: &heapReleased}

	heapSys := float64(rtm.HeapSys)
	a.Metrics["HeapSys"] = models.Metrics{ID: "HeapSys", MType: models.Gauge, Value: &heapSys}

	lastGC := float64(rtm.LastGC)
	a.Metrics["LastGC"] = models.Metrics{ID: "LastGC", MType: models.Gauge, Value: &lastGC}

	lookups := float64(rtm.Lookups)
	a.Metrics["Lookups"] = models.Metrics{ID: "Lookups", MType: models.Gauge, Value: &lookups}

	mCacheInuse := float64(rtm.MCacheInuse)
	a.Metrics["MCacheInuse"] = models.Metrics{ID: "MCacheInuse", MType: models.Gauge, Value: &mCacheInuse}

	mCacheSys := float64(rtm.MCacheSys)
	a.Metrics["MCacheSys"] = models.Metrics{ID: "MCacheSys", MType: models.Gauge, Value: &mCacheSys}

	mSpanInuse := float64(rtm.MSpanInuse)
	a.Metrics["MSpanInuse"] = models.Metrics{ID: "MSpanInuse", MType: models.Gauge, Value: &mSpanInuse}

	mSpanSys := float64(rtm.MSpanSys)
	a.Metrics["MSpanSys"] = models.Metrics{ID: "MSpanSys", MType: models.Gauge, Value: &mSpanSys}

	mallocs := float64(rtm.Mallocs)
	a.Metrics["Mallocs"] = models.Metrics{ID: "Mallocs", MType: models.Gauge, Value: &mallocs}

	nextGC := float64(rtm.NextGC)
	a.Metrics["NextGC"] = models.Metrics{ID: "NextGC", MType: models.Gauge, Value: &nextGC}

	numForcedGC := float64(rtm.NumForcedGC)
	a.Metrics["NumForcedGC"] = models.Metrics{ID: "NumForcedGC", MType: models.Gauge, Value: &numForcedGC}

	numGC := float64(rtm.NumGC)
	a.Metrics["NumGC"] = models.Metrics{ID: "NumGC", MType: models.Gauge, Value: &numGC}

	otherSys := float64(rtm.OtherSys)
	a.Metrics["OtherSys"] = models.Metrics{ID: "OtherSys", MType: models.Gauge, Value: &otherSys}

	pauseTotalNs := float64(rtm.PauseTotalNs)
	a.Metrics["PauseTotalNs"] = models.Metrics{ID: "PauseTotalNs", MType: models.Gauge, Value: &pauseTotalNs}

	stackInuse := float64(rtm.StackInuse)
	a.Metrics["StackInuse"] = models.Metrics{ID: "StackInuse", MType: models.Gauge, Value: &stackInuse}

	stackSys := float64(rtm.StackSys)
	a.Metrics["StackSys"] = models.Metrics{ID: "StackSys", MType: models.Gauge, Value: &stackSys}

	sys := float64(rtm.Sys)
	a.Metrics["Sys"] = models.Metrics{ID: "Sys", MType: models.Gauge, Value: &sys}

	totalAlloc := float64(rtm.TotalAlloc)
	a.Metrics["TotalAlloc"] = models.Metrics{ID: "TotalAlloc", MType: models.Gauge, Value: &totalAlloc}

	randomValue := rand.Float64()
	a.Metrics["RandomValue"] = models.Metrics{ID: "RandomValue", MType: models.Gauge, Value: &randomValue}

	pollCount := a.PollCount
	a.Metrics["PollCount"] = models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &pollCount}

	a.PollCount++
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
