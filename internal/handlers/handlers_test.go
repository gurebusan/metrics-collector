package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/models"
	me "github.com/zetcan333/metrics-collector/pkg/myerrors"
)

// Мок для ServerUseCase
// Мок для ServerUseCase
type mockUseCase struct {
	updateMetricFunc  func(metricType, metricName, metricValue string) error
	getValueFunc      func(metricType, metricName string) (string, error)
	updateJSONFunc    func(metric models.Metrics) (models.Metrics, error)
	getJSONFunc       func(metric models.Metrics) (models.Metrics, error)
	getAllMetricsFunc func() (string, error)
}

func (m *mockUseCase) UpdateMetric(metricType, metricName, metricValue string) error {
	return m.updateMetricFunc(metricType, metricName, metricValue)
}

func (m *mockUseCase) GetValue(metricType, metricName string) (string, error) {
	return m.getValueFunc(metricType, metricName)
}

func (m *mockUseCase) UpdateJSON(metric models.Metrics) (models.Metrics, error) {
	return m.updateJSONFunc(metric)
}

func (m *mockUseCase) GetJSON(metric models.Metrics) (models.Metrics, error) {
	return m.getJSONFunc(metric)
}

func (m *mockUseCase) GetAllMetric() (string, error) {
	return m.getAllMetricsFunc()
}

func TestUpdateHandler(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		mockFunc     func(metricType, metricName, metricValue string) error
		expectedCode int
	}{
		{
			name: "Valid gauge metric",
			path: "/update/gauge/testGauge/123.45",
			mockFunc: func(metricType, metricName, metricValue string) error {
				return nil
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Valid counter metric",
			path: "/update/counter/testCounter/100",
			mockFunc: func(metricType, metricName, metricValue string) error {
				return nil
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid metric type",
			path: "/update/unknown/testMetric/123",
			mockFunc: func(metricType, metricName, metricValue string) error {
				return me.ErrInvalidMetricType
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid gauge value",
			path: "/update/gauge/testGauge/invalid",
			mockFunc: func(metricType, metricName, metricValue string) error {
				return me.ErrInvalidGaugeValue
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Invalid counter value",
			path: "/update/counter/testCounter/invalid",
			mockFunc: func(metricType, metricName, metricValue string) error {
				return me.ErrInvalidCounterValue
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "Empty metric name",
			path: "/update/gauge//123.45",
			mockFunc: func(metricType, metricName, metricValue string) error {
				return nil
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name: "Invalid path",
			path: "/update/gauge/testGauge",
			mockFunc: func(metricType, metricName, metricValue string) error {
				return nil
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём мок usecase
			mockUsecase := &mockUseCase{
				updateMetricFunc: tt.mockFunc,
			}

			// Создаём хэндлер
			handler := handlers.NewServerHandler(mockUsecase)

			// Создаём роутер chi
			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler)

			// Создаём тестовый HTTP-запрос
			req, err := http.NewRequest(http.MethodPost, tt.path, nil)
			require.NoError(t, err, "failed to create request")

			// Создаём ResponseRecorder для получения ответа
			rr := httptest.NewRecorder()

			// Вызываем роутер
			r.ServeHTTP(rr, req)

			// Проверяем HTTP-код ответа
			assert.Equal(t, tt.expectedCode, rr.Code, "unexpected status code")
		})
	}
}

func TestGetValueHandler(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		mockFunc     func(metricType, metricName string) (string, error)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Valid gauge metric",
			path: "/value/gauge/Alloc",
			mockFunc: func(metricType, metricName string) (string, error) {
				return "123.45", nil
			},
			expectedCode: http.StatusOK,
			expectedBody: "123.45",
		},
		{
			name: "Valid counter metric",
			path: "/value/counter/PollCount",
			mockFunc: func(metricType, metricName string) (string, error) {
				return "10", nil
			},
			expectedCode: http.StatusOK,
			expectedBody: "10",
		},
		{
			name: "Unknown metric type",
			path: "/value/unknown/Test",
			mockFunc: func(metricType, metricName string) (string, error) {
				return "", me.ErrInvalidMetricType
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "invalid metric type\n",
		},
		{
			name: "Non-existent metric",
			path: "/value/gauge/UnknownMetric",
			mockFunc: func(metricType, metricName string) (string, error) {
				return "", me.ErrMetricNotFound
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "metric not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём мок usecase
			mockUsecase := &mockUseCase{
				getValueFunc: tt.mockFunc,
			}

			// Создаём хэндлер
			handler := handlers.NewServerHandler(mockUsecase)

			// Создаём роутер chi
			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", handler.GetValueHandler)

			// Создаём тестовый HTTP-запрос
			req, err := http.NewRequest(http.MethodGet, tt.path, nil)
			require.NoError(t, err, "failed to create request")

			// Создаём ResponseRecorder для получения ответа
			rr := httptest.NewRecorder()

			// Вызываем роутер
			r.ServeHTTP(rr, req)

			// Проверяем HTTP-код ответа
			assert.Equal(t, tt.expectedCode, rr.Code, "unexpected status code")

			// Проверяем тело ответа
			assert.Equal(t, tt.expectedBody, rr.Body.String(), "unexpected response body")
		})
	}
}
func TestUpdateJSONHandler(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  models.Metrics
		mockFunc     func(metric models.Metrics) (models.Metrics, error)
		expectedCode int
	}{
		{
			name: "Valid JSON update",
			requestBody: models.Metrics{
				ID:    "testGauge",
				MType: "gauge",
				Value: func() *float64 { v := 123.45; return &v }(),
			},
			mockFunc: func(metric models.Metrics) (models.Metrics, error) {
				return metric, nil
			},
			expectedCode: http.StatusOK,
		},
		{
			name:        "Invalid JSON format",
			requestBody: models.Metrics{},
			mockFunc: func(metric models.Metrics) (models.Metrics, error) {
				return models.Metrics{}, me.ErrInvalidMetricType
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockUseCase{
				updateJSONFunc: tt.mockFunc,
			}
			handler := handlers.NewServerHandler(mockUsecase)
			r := chi.NewRouter()
			r.Post("/update/", handler.UpdateJSONHandler)

			requestBody, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewReader(requestBody))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}

func TestGetJSONHandler(t *testing.T) {
	tests := []struct {
		name         string
		input        models.Metrics
		mockFunc     func(models.Metrics) (models.Metrics, error)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Valid gauge metric",
			input: models.Metrics{
				ID:    "Alloc",
				MType: "gauge",
			},
			mockFunc: func(m models.Metrics) (models.Metrics, error) {
				m.Value = new(float64)
				*m.Value = 123.45
				return m, nil
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"Alloc","type":"gauge","value":123.45}`,
		},
		{
			name: "Valid counter metric",
			input: models.Metrics{
				ID:    "PollCount",
				MType: "counter",
			},
			mockFunc: func(m models.Metrics) (models.Metrics, error) {
				m.Delta = new(int64)
				*m.Delta = 10
				return m, nil
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"PollCount","type":"counter","delta":10}`,
		},
		{
			name: "Metric not found",
			input: models.Metrics{
				ID:    "Unknown",
				MType: "gauge",
			},
			mockFunc: func(m models.Metrics) (models.Metrics, error) {
				return models.Metrics{}, me.ErrMetricNotFound
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "metric not found\n",
		},
		{
			name: "Invalid metric type",
			input: models.Metrics{
				ID:    "Invalid",
				MType: "unknown",
			},
			mockFunc: func(m models.Metrics) (models.Metrics, error) {
				return models.Metrics{}, me.ErrInvalidMetricType
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "invalid metric type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockUseCase{
				getJSONFunc: tt.mockFunc,
			}

			handler := handlers.NewServerHandler(mockUsecase)

			r := chi.NewRouter()
			r.Post("/value/", handler.GetJSONHandler)

			body, _ := json.Marshal(tt.input)
			req, err := http.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(body))
			require.NoError(t, err, "failed to create request")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code, "unexpected status code")
			if tt.expectedCode == http.StatusOK {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String(), "unexpected response body")
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String(), "unexpected response body")
			}
		})
	}
}

func TestGetAllMetricsHandler(t *testing.T) {
	tests := []struct {
		name         string
		mockFunc     func() (string, error)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Valid metrics page",
			mockFunc: func() (string, error) {
				return "<html>Metrics</html>", nil
			},
			expectedCode: http.StatusOK,
			expectedBody: "<html>Metrics</html>",
		},
		{
			name: "Internal server error",
			mockFunc: func() (string, error) {
				return "", errors.New("internal error")
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockUseCase{
				getAllMetricsFunc: tt.mockFunc,
			}
			handler := handlers.NewServerHandler(mockUsecase)
			r := chi.NewRouter()
			r.Get("/metrics", handler.GetAllMetricsHandler)

			req, err := http.NewRequest(http.MethodGet, "/metrics", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
		})
	}
}
