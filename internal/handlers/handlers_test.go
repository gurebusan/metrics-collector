package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zetcan333/metrics-collector/internal/handlers"
)

type mockStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func (m *mockStorage) UpdateGauge(name string, value float64) {
	m.gauge[name] = value
}
func (m *mockStorage) UpdateCounter(name string, value int64) {
	m.counter[name] += value
}

func (m *mockStorage) GetGauge(name string) (float64, bool) {
	val, ok := m.gauge[name]
	return val, ok
}
func (m *mockStorage) GetCounter(name string) (int64, bool) {
	val, ok := m.counter[name]
	return val, ok
}
func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		expectedCode    int
		expectedGauge   map[string]float64
		expectedCounter map[string]int64
	}{
		{
			name:            "Valid gauge metric",
			path:            "/update/gauge/testGauge/123.45",
			expectedCode:    http.StatusOK,
			expectedGauge:   map[string]float64{"testGauge": 123.45},
			expectedCounter: map[string]int64{},
		},
		{
			name:            "Valid counter metric",
			path:            "/update/counter/testCounter/100",
			expectedCode:    http.StatusOK,
			expectedGauge:   map[string]float64{},
			expectedCounter: map[string]int64{"testCounter": 100},
		},
		{
			name:            "Invalid metric type",
			path:            "/update/unknown/testMetric/123",
			expectedCode:    http.StatusBadRequest,
			expectedGauge:   map[string]float64{},
			expectedCounter: map[string]int64{},
		},
		{
			name:            "Invalid gauge value",
			path:            "/update/gauge/testGauge/invalid",
			expectedCode:    http.StatusBadRequest,
			expectedGauge:   map[string]float64{},
			expectedCounter: map[string]int64{},
		},
		{
			name:            "Invalid counter value",
			path:            "/update/counter/testCounter/invalid",
			expectedCode:    http.StatusBadRequest,
			expectedGauge:   map[string]float64{},
			expectedCounter: map[string]int64{},
		},
		{
			name:            "Empty metric name",
			path:            "/update/gauge//123.45",
			expectedCode:    http.StatusNotFound,
			expectedGauge:   map[string]float64{},
			expectedCounter: map[string]int64{},
		},
		{
			name:            "Invalid path",
			path:            "/update/gauge/testGauge",
			expectedCode:    http.StatusNotFound,
			expectedGauge:   map[string]float64{},
			expectedCounter: map[string]int64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockStorage{
				gauge:   make(map[string]float64),
				counter: make(map[string]int64),
			}
			req, err := http.NewRequest(http.MethodPost, tt.path, nil)
			require.NoError(t, err, "failed to create request")

			//Создаём новый рекордер
			w := httptest.NewRecorder()

			// Создаём обработчик и вызываем его
			handler := handlers.UpdateHandler(updater)
			handler.ServeHTTP(w, req)

			// Проверяем HTTP-код ответа
			assert.Equal(t, tt.expectedCode, w.Code, "unexpected status code")

			// Проверяем обновлённые метрики
			for key, expectedValue := range tt.expectedGauge {
				assert.Contains(t, updater.gauge, key, "gauge metric not found")
				assert.Equal(t, expectedValue, updater.gauge[key], "unexpected gauge value")
			}
			for key, expectedValue := range tt.expectedCounter {
				assert.Contains(t, updater.counter, key, "counter metric not found")
				assert.Equal(t, expectedValue, updater.counter[key], "unexpected counter value")
			}
		})
	}
}

func TestGetValueHandler(t *testing.T) {
	getter := &mockStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
	getter.UpdateGauge("Alloc", 123.45)
	getter.UpdateCounter("PollCount", 10)

	tests := []struct {
		name         string
		path         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Valid gauge metric",
			path:         "/value/gauge/Alloc",
			expectedCode: http.StatusOK,
			expectedBody: "123.45",
		},
		{
			name:         "Valid counter metric",
			path:         "/value/counter/PollCount",
			expectedCode: http.StatusOK,
			expectedBody: "10",
		},
		{
			name:         "Unknown metric type",
			path:         "/value/unknown/Test",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Invalid metric type\n",
		},
		{
			name:         "Non-existent metric",
			path:         "/value/gauge/UnknownMetric",
			expectedCode: http.StatusNotFound,
			expectedBody: "Metric not found\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём тестовый HTTP-запрос
			req, err := http.NewRequest("GET", tt.path, nil)
			assert.NoError(t, err)

			// Создаём ResponseRecorder для получения ответа
			rr := httptest.NewRecorder()

			// Вызов хэндлера напрямую
			handler := handlers.GetValueHandler(getter)
			handler.ServeHTTP(rr, req)

			// Проверяем код ответа
			assert.Equal(t, tt.expectedCode, rr.Code, "unexpected status code")

			// Проверяем тело ответа
			assert.Equal(t, tt.expectedBody, rr.Body.String(), "unexpected response body")
		})
	}
}
