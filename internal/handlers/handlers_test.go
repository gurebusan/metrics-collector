package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	me "github.com/zetcan333/metrics-collector/pkg/myerrors"
)

// Мок для ServerUseCase
type mockUseCase struct {
	updateMetricFunc func(metricType, metricName, metricValue string) error
	getValueFunc     func(metricType, metricName string) (string, error)
}

func (m *mockUseCase) UpdateMetric(metricType, metricName, metricValue string) error {
	return m.updateMetricFunc(metricType, metricName, metricValue)
}

func (m *mockUseCase) GetValue(metricType, metricName string) (string, error) {
	return m.getValueFunc(metricType, metricName)
}

func (m *mockUseCase) GetAllMetric() (string, error) {
	return "", nil
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
			r.Post("/update/{type}/{name}/{value}", handler.UpdateHandle)

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
