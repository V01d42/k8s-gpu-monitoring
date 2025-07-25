package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"k8s-gpu-monitoring/internal/models"
	"k8s-gpu-monitoring/internal/prometheus"
)

// PrometheusClient interface for testing
type PrometheusClient interface {
	GetGPUMetrics(ctx context.Context) ([]models.GPUMetrics, error)
	GetGPUNodes(ctx context.Context) ([]models.GPUNode, error)
	Query(ctx context.Context, query string) (*prometheus.PrometheusResponse, error)
}

type mockPrometheusClient struct {
	shouldReturnError bool
}

func (m *mockPrometheusClient) GetGPUMetrics(ctx context.Context) ([]models.GPUMetrics, error) {
	if m.shouldReturnError {
		return nil, errors.New("mock prometheus error")
	}

	return []models.GPUMetrics{
		{
			NodeName:          "node1",
			GPUIndex:          0,
			GPUName:           "NVIDIA Tesla V100",
			Utilization:       75.5,
			MemoryUsed:        8.0,
			MemoryTotal:       16.0,
			MemoryFree:        8.0,
			MemoryUtilization: 50.0,
			Temperature:       65.0,
			PowerDraw:         250.0,
			PowerLimit:        300.0,
		},
	}, nil
}

func (m *mockPrometheusClient) GetGPUNodes(ctx context.Context) ([]models.GPUNode, error) {
	if m.shouldReturnError {
		return nil, errors.New("mock prometheus error")
	}

	return []models.GPUNode{
		{
			NodeName:  "node1",
			GPUCount:  2,
			GPUModels: []string{"NVIDIA Tesla V100"},
		},
	}, nil
}

func (m *mockPrometheusClient) Query(ctx context.Context, query string) (*prometheus.PrometheusResponse, error) {
	if m.shouldReturnError {
		return nil, errors.New("mock prometheus error")
	}

	return &prometheus.PrometheusResponse{
		Status: "success",
		Data: struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		}{
			ResultType: "vector",
			Result: []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			}{
				{
					Metric: map[string]string{
						"node": "node1",
						"gpu":  "0",
					},
					Value: []interface{}{
						1234567890.0,
						"75.5",
					},
				},
			},
		},
	}, nil
}

// GPUHandlerWrapper for testing with interface
type GPUHandlerWrapper struct {
	promClient PrometheusClient
}

func NewGPUHandlerWrapper(promClient PrometheusClient) *GPUHandlerWrapper {
	return &GPUHandlerWrapper{
		promClient: promClient,
	}
}

func (h *GPUHandlerWrapper) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *GPUHandlerWrapper) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := h.promClient.Query(ctx, "up")
	if err != nil {
		response := models.APIResponse{
			Success: false,
			Error:   "Prometheus connection failed",
		}
		h.writeJSONResponse(w, http.StatusServiceUnavailable, response)
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Service is healthy",
	}
	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *GPUHandlerWrapper) GetGPUMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	metrics, err := h.promClient.GetGPUMetrics(ctx)
	if err != nil {
		response := models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve GPU metrics",
		}
		h.writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    metrics,
		Message: "GPU metrics retrieved successfully",
	}
	h.writeJSONResponse(w, http.StatusOK, response)
}

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		mockError    bool
		expectedCode int
	}{
		{
			name:         "successful health check",
			mockError:    false,
			expectedCode: http.StatusOK,
		},
		{
			name:         "prometheus connection error",
			mockError:    true,
			expectedCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPrometheusClient{shouldReturnError: tt.mockError}
			handler := NewGPUHandlerWrapper(mockClient)

			req := httptest.NewRequest("GET", "/api/health", nil)
			w := httptest.NewRecorder()

			handler.HealthCheck(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var response models.APIResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if tt.mockError && response.Success {
				t.Error("expected success to be false for error case")
			}
			if !tt.mockError && !response.Success {
				t.Error("expected success to be true for success case")
			}
		})
	}
}

func TestGetGPUMetrics(t *testing.T) {
	tests := []struct {
		name         string
		mockError    bool
		expectedCode int
		checkData    bool
	}{
		{
			name:         "successful metrics retrieval",
			mockError:    false,
			expectedCode: http.StatusOK,
			checkData:    true,
		},
		{
			name:         "prometheus error",
			mockError:    true,
			expectedCode: http.StatusInternalServerError,
			checkData:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockPrometheusClient{shouldReturnError: tt.mockError}
			handler := NewGPUHandlerWrapper(mockClient)

			req := httptest.NewRequest("GET", "/api/v1/gpu/metrics", nil)
			w := httptest.NewRecorder()

			handler.GetGPUMetrics(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var response models.APIResponse
			err := json.NewDecoder(w.Body).Decode(&response)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if tt.checkData {
				if !response.Success {
					t.Error("expected success to be true")
				}
				if response.Data == nil {
					t.Error("expected data to be present")
				}
			}
		})
	}
}
