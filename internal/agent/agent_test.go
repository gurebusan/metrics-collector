package agent_test

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zetcan333/metrics-collector/internal/agent"
	"github.com/zetcan333/metrics-collector/internal/models"
)

func TestCollectMetrics(t *testing.T) {
	a := agent.NewAgent("http://localhost:8080", 2*time.Second, 10*time.Second)
	a.CollectMetrics()

	// Проверяем, что метрики заполнены
	assert.NotEmpty(t, a.Metrics, "Metrics не должны быть пустыми")
	assert.Contains(t, a.Metrics, "PollCount", "Должен быть счетчик PollCount")
	assert.Contains(t, a.Metrics, "RandomValue", "Должен быть RandomValue")

	// Проверяем, что RandomValue находится в пределах от 0 до 1
	randomMetric := a.Metrics["RandomValue"]
	assert.NotNil(t, randomMetric.Value, "RandomValue должен быть задан")
	assert.GreaterOrEqual(t, *randomMetric.Value, 0.0, "RandomValue должен быть >= 0")
	assert.LessOrEqual(t, *randomMetric.Value, 1.0, "RandomValue должен быть <= 1")

	// Проверяем, что PollCount увеличился
	prevPollCount := *a.Metrics["PollCount"].Delta
	a.CollectMetrics()
	assert.Equal(t, prevPollCount+1, *a.Metrics["PollCount"].Delta, "PollCount должен увеличиваться")
}

func TestSendMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/update", "Неправильный URL")

		// Проверяем заголовок Content-Encoding
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"), "Данные должны быть сжаты в gzip")

		// Декодируем тело запроса
		gzReader, err := gzip.NewReader(r.Body)
		assert.NoError(t, err, "Ошибка при создании gzip.Reader")
		defer gzReader.Close()

		var metric models.Metrics
		err = json.NewDecoder(gzReader).Decode(&metric)
		assert.NoError(t, err, "Ошибка при декодировании JSON")
		assert.NotEmpty(t, metric.ID, "ID метрики не должен быть пустым")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	a := agent.NewAgent(server.URL, 2*time.Millisecond, 10*time.Millisecond)
	value := 123.45
	count := int64(100)
	a.Metrics["TestGauge"] = models.Metrics{ID: "TestGauge", MType: models.Gauge, Value: &value}
	a.Metrics["TestCounter"] = models.Metrics{ID: "TestCounter", MType: models.Counter, Delta: &count}

	err := a.SendMetrics()
	assert.NoError(t, err, "Отправка метрик должна выполняться без ошибок")
}

func TestSendMetricsBatch(t *testing.T) {

	expectedGauge := 123.45
	expectedCounter := int64(100)
	expectedMetrics := []models.Metrics{
		{ID: "TestGauge", MType: models.Gauge, Value: &expectedGauge},
		{ID: "TestCounter", MType: models.Counter, Delta: &expectedCounter},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		assert.Equal(t, http.MethodPost, r.Method, "Должен быть POST-запрос")
		assert.Equal(t, "/updates/", r.URL.Path, "Неправильный путь")

		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"), "Должно быть gzip-сжатие")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Должен быть JSON")

		gzReader, err := gzip.NewReader(r.Body)
		require.NoError(t, err, "Ошибка распаковки gzip")
		defer gzReader.Close()

		var receivedMetrics []models.Metrics
		err = json.NewDecoder(gzReader).Decode(&receivedMetrics)
		require.NoError(t, err, "Ошибка декодирования JSON")

		require.Len(t, receivedMetrics, 2, "Должно быть 2 метрики")
		assert.Equal(t, expectedMetrics, receivedMetrics, "Метрики не совпадают")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	a := agent.NewAgent(server.URL, time.Minute, time.Minute)

	a.Metrics = make(map[string]models.Metrics)
	a.Metrics["TestGauge"] = expectedMetrics[0]
	a.Metrics["TestCounter"] = expectedMetrics[1]

	err := a.SendMetricsBatch()
	assert.NoError(t, err, "Не должно быть ошибки")
}
