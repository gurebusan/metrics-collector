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
