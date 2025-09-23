package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"k8s-gpu-monitoring/internal/models"
	"k8s-gpu-monitoring/internal/prometheus"
	"k8s-gpu-monitoring/internal/timeutil"
)

// GPUHandler handles GPU-related HTTP requests with Prometheus backend.
type GPUHandler struct {
	promClient *prometheus.Client
}

// NewGPUHandler creates a new GPU handler with the provided Prometheus client.
func NewGPUHandler(promClient *prometheus.Client) *GPUHandler {
	return &GPUHandler{
		promClient: promClient,
	}
}

// writeJSONResponse writes a JSON response with proper headers.
func (h *GPUHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// writeErrorResponse writes a standardized error response.
func (h *GPUHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := models.APIResponse{
		Success: false,
		Error:   message,
	}
	h.writeJSONResponse(w, statusCode, response)
}

// GetGPUMetrics handles GET /api/v1/gpu/metrics - returns comprehensive GPU metrics.
func (h *GPUHandler) GetGPUMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	metrics, err := h.promClient.GetGPUMetrics(ctx)
	if err != nil {
		log.Printf("Error getting GPU metrics: %v", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve GPU metrics")
		return
	}

	response := models.APIResponse{
		Success: true,
		Data:    metrics,
		Message: "GPU metrics retrieved successfully",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// HealthCheck handles GET /api/healthz - verifies service and Prometheus connectivity.
func (h *GPUHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Verify Prometheus server connectivity
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err := h.promClient.Query(ctx, "up")
	if err != nil {
		log.Printf("Health check failed: %v", err)
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Prometheus connection failed")
		return
	}

	response := models.APIResponse{
		Success: true,
		Message: "Service is healthy",
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": timeutil.NowJST(),
			"version":   "1.0.0",
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}
