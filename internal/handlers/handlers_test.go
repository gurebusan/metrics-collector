package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zetcan333/metrics-collector/internal/handlers"
	"github.com/zetcan333/metrics-collector/internal/handlers/mocks"
	"github.com/zetcan333/metrics-collector/internal/lib/zapdiscard"
	"github.com/zetcan333/metrics-collector/internal/models"
	"github.com/zetcan333/metrics-collector/pkg/myerrors"
)

func TestUpdateMetric(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		mockSetup    func(*mocks.ServerUseCase)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success gauge update",
			path: "/update/gauge/testGauge/123.45",
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("UpdateMetric", "gauge", "testGauge", "123.45").Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid metric type",
			path: "/update/invalid/test/123",
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("UpdateMetric", "invalid", "test", "123").Return(myerrors.ErrInvalidMetricType)
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "invalid metric type\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mocks.ServerUseCase{}
			tt.mockSetup(mockUsecase)

			handler := handlers.NewServerHandler(zapdiscard.NewDiscardLogger(), mockUsecase)
			r := chi.NewRouter()
			r.Post("/update/{type}/{name}/{value}", handler.UpdateMetric)

			req, err := http.NewRequest(http.MethodPost, tt.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestGetMetric(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		mockSetup    func(*mocks.ServerUseCase)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success get gauge",
			path: "/value/gauge/testGauge",
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("GetMetric", "gauge", "testGauge").Return("123.45", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "123.45",
		},
		{
			name: "Metric not found",
			path: "/value/gauge/notfound",
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("GetMetric", "gauge", "notfound").Return("", myerrors.ErrMetricNotFound)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "metric not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mocks.ServerUseCase{}
			tt.mockSetup(mockUsecase)

			handler := handlers.NewServerHandler(zapdiscard.NewDiscardLogger(), mockUsecase)
			r := chi.NewRouter()
			r.Get("/value/{type}/{name}", handler.GetMetric)

			req, err := http.NewRequest(http.MethodGet, tt.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestUpdateViaModel(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  models.Metrics
		mockSetup    func(*mocks.ServerUseCase)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success via model update",
			requestBody: models.Metrics{
				ID:    "testGauge",
				MType: "gauge",
				Value: func() *float64 { v := 123.45; return &v }(),
			},
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("UpdateViaModel", mock.AnythingOfType("models.Metrics")).
					Return(models.Metrics{
						ID:    "testGauge",
						MType: "gauge",
						Value: func() *float64 { v := 123.45; return &v }(),
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"testGauge","type":"gauge","value":123.45}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mocks.ServerUseCase{}
			tt.mockSetup(mockUsecase)

			handler := handlers.NewServerHandler(zapdiscard.NewDiscardLogger(), mockUsecase)
			r := chi.NewRouter()
			r.Post("/update/", handler.UpdateViaModel)

			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			}
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestGetViaModel(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  models.Metrics
		mockSetup    func(*mocks.ServerUseCase)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success get metric via model",
			requestBody: models.Metrics{
				ID:    "testGauge",
				MType: "gauge",
			},
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("GetViaModel", mock.AnythingOfType("models.Metrics")).
					Return(models.Metrics{
						ID:    "testGauge",
						MType: "gauge",
						Value: func() *float64 { v := 123.45; return &v }(),
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"id":"testGauge","type":"gauge","value":123.45}`,
		},
		{
			name: "Metric not found",
			requestBody: models.Metrics{
				ID:    "notfound",
				MType: "gauge",
			},
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("GetViaModel", mock.AnythingOfType("models.Metrics")).
					Return(models.Metrics{}, myerrors.ErrMetricNotFound)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: `{"error":"metric not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mocks.ServerUseCase{}
			tt.mockSetup(mockUsecase)

			handler := handlers.NewServerHandler(zapdiscard.NewDiscardLogger(), mockUsecase)
			r := chi.NewRouter()
			r.Post("/value/", handler.GetViaModel)

			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			}
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mocks.ServerUseCase)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success get all metrics",
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("GetAllMetrics").Return("<html>metrics</html>", nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: "<html>metrics</html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mocks.ServerUseCase{}
			tt.mockSetup(mockUsecase)

			handler := handlers.NewServerHandler(zapdiscard.NewDiscardLogger(), mockUsecase)
			r := chi.NewRouter()
			r.Get("/", handler.GetAllMetrics)

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestUpdateMetricsWithBatch(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  []models.Metrics
		mockSetup    func(*mocks.ServerUseCase)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success update metrics with batch",
			requestBody: []models.Metrics{
				{
					ID:    "testGauge",
					MType: "gauge",
					Value: func() *float64 { v := 123.45; return &v }(),
				},
				{
					ID:    "testCounter",
					MType: "counter",
					Delta: func() *int64 { v := int64(10); return &v }(),
				},
			},
			mockSetup: func(m *mocks.ServerUseCase) {
				m.On("UpdateMetricsWithBatch", mock.AnythingOfType("[]models.Metrics")).
					Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"status":"batch updated successfully"}`,
		},
		{
			name:        "Empty batch",
			requestBody: []models.Metrics{},
			mockSetup: func(m *mocks.ServerUseCase) {
				// No expectations as validation fails before usecase is called
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"empty batch"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mocks.ServerUseCase{}
			tt.mockSetup(mockUsecase)

			handler := handlers.NewServerHandler(zapdiscard.NewDiscardLogger(), mockUsecase)
			r := chi.NewRouter()
			r.Post("/updates/", handler.UpdateMetricsWithBatch)

			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			}
			mockUsecase.AssertExpectations(t)
		})
	}
}
