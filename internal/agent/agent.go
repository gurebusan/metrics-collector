package agent

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Структура агента
type Agent struct {
	ServerURL      string
	PollInterval   time.Duration
	ReportInterval time.Duration
	GaugeMetric    map[string]float64
	CounterMetric  map[string]int64
	client         http.Client
	sync.RWMutex
}

// Конструктор агента
func NewAgent(serverURL string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		ServerURL:      serverURL,
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		GaugeMetric:    make(map[string]float64),
		CounterMetric:  make(map[string]int64),
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
	//Отпправляем gauge-метрики
	for name, value := range a.GaugeMetric {
		destination := fmt.Sprintf("%s/update/gauge/%s/%v", a.ServerURL, name, value)
		resp, err := a.client.Post(destination, "text/plain", nil)
		if err != nil {
			return fmt.Errorf("failed to send gauge metric %s: %v", name, value)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d for gauge metric %s", resp.StatusCode, name)
		}
	}
	// Отпраялем counter-метрики
	for name, value := range a.CounterMetric {
		destination := fmt.Sprintf("%s/update/counter/%s/%v", a.ServerURL, name, value)
		resp, err := http.Post(destination, "text/plain", nil)
		if err != nil {
			return fmt.Errorf("failed to send counter metric %s: %v", name, value)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d for counter metric %s", resp.StatusCode, name)
		}
	}
	return nil
}

// Сбор метрик из rutime
func (a *Agent) CollectMetrics() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	a.Lock()
	defer a.Unlock()
	// Обновляем gauge-метрики
	a.GaugeMetric["Alloc"] = float64(rtm.Alloc)
	a.GaugeMetric["BuckHashSys"] = float64(rtm.BuckHashSys)
	a.GaugeMetric["Frees"] = float64(rtm.Frees)
	a.GaugeMetric["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	a.GaugeMetric["GCSys"] = float64(rtm.GCSys)
	a.GaugeMetric["HeapAlloc"] = float64(rtm.HeapAlloc)
	a.GaugeMetric["HeapIdle"] = float64(rtm.HeapIdle)
	a.GaugeMetric["HeapInuse"] = float64(rtm.HeapInuse)
	a.GaugeMetric["HeapObjects"] = float64(rtm.HeapObjects)
	a.GaugeMetric["HeapReleased"] = float64(rtm.HeapReleased)
	a.GaugeMetric["HeapSys"] = float64(rtm.HeapSys)
	a.GaugeMetric["LastGC"] = float64(rtm.LastGC)
	a.GaugeMetric["Lookups"] = float64(rtm.Lookups)
	a.GaugeMetric["MCacheInuse"] = float64(rtm.MCacheInuse)
	a.GaugeMetric["MCacheSys"] = float64(rtm.MCacheSys)
	a.GaugeMetric["MSpanInuse"] = float64(rtm.MSpanInuse)
	a.GaugeMetric["MSpanSys"] = float64(rtm.MSpanSys)
	a.GaugeMetric["Mallocs"] = float64(rtm.Mallocs)
	a.GaugeMetric["NextGC"] = float64(rtm.NextGC)
	a.GaugeMetric["NumForcedGC"] = float64(rtm.NumForcedGC)
	a.GaugeMetric["NumGC"] = float64(rtm.NumGC)
	a.GaugeMetric["OtherSys"] = float64(rtm.OtherSys)
	a.GaugeMetric["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	a.GaugeMetric["StackInuse"] = float64(rtm.StackInuse)
	a.GaugeMetric["StackSys"] = float64(rtm.StackSys)
	a.GaugeMetric["Sys"] = float64(rtm.Sys)
	a.GaugeMetric["TotalAlloc"] = float64(rtm.TotalAlloc)
	a.GaugeMetric["RandomValue"] = rand.Float64()
	// Обновляем counter-метрики
	a.CounterMetric["PollCount"]++
}
