package agent_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zetcan333/metrics-collector/internal/agent"
)

func TestCollectMetrics(t *testing.T) {
	//Инициализация агента
	a := agent.NewAgent("http://localhost:8080", "text/plain", 2, 10)

	//Вызывем метод сбора метрик
	a.CollectMetrics()

	// Проверяем, что метрики заполнены
	assert.NotEmpty(t, a.GaugeMetric, "GaugeMetric не должен быть пустым")
	assert.NotEmpty(t, a.CounterMetric, "CounterMetric не должен быть пустым")
	assert.Contains(t, a.CounterMetric, "PollCount", "Должен быть счетчик PollCount")

	// Проверяем, что RandomValue обновился и находится в пределах от 0 до 1
	randomValue := a.GaugeMetric["RandomValue"]
	assert.GreaterOrEqual(t, randomValue, 0.0, "RandomValue должен быть >= 0")
	assert.LessOrEqual(t, randomValue, 1.0, "RandomValue должен быть <= 1")

	// Проверяем, что PollCount увеличился
	prevPollCount := a.CounterMetric["PollCount"]
	a.CollectMetrics()
	assert.Equal(t, prevPollCount+1, a.CounterMetric["PollCount"], "PollCount должен увеличиваться")
}

func TestSendMetrics(t *testing.T) {
	//Создаем тестовый HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/update/", "Неправильный URL")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	a := agent.NewAgent(server.URL, "text/plain", 2*time.Millisecond, 10*time.Microsecond)
	a.GaugeMetric["TestGauge"] = 123.45
	a.CounterMetric["TestCounter"] = 100

	err := a.SendMetrics()
	assert.NoError(t, err, "Отправка метрик должна выполняться без ошибок")
}

func TestStart(t *testing.T) {
	a := agent.NewAgent("http://localhost:8080", "text/plain", 1*time.Millisecond, 2*time.Millisecond)

	// Запускаем агент
	go a.Start()

	// Даем время на выполнение (можно уменьшить в реальном коде)
	time.Sleep(10 * time.Millisecond)

	// Проверяем, что метрики обновились
	assert.NotEmpty(t, a.GaugeMetric, "GaugeMetric должен обновляться")
	assert.NotEmpty(t, a.CounterMetric, "CounterMetric должен обновляться")
}
