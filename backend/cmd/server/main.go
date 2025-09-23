package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s-gpu-monitoring/internal/handlers"
	"k8s-gpu-monitoring/internal/middleware"
	"k8s-gpu-monitoring/internal/prometheus"
)

// main starts the GPU monitoring API server with graceful shutdown support.
func main() {
	// Load configuration from environment variables
	prometheusURL := getEnv("PROMETHEUS_URL", "http://localhost:9090")
	port := getEnv("PORT", "8080")

	log.Printf("Starting GPU Monitoring API Server...")
	log.Printf("Prometheus URL: %s", prometheusURL)
	log.Printf("Server Port: %s", port)

	// Initialize Prometheus client
	promClient := prometheus.NewClient(prometheusURL)

	// Initialize handlers
	gpuHandler := handlers.NewGPUHandler(promClient)

	// Use Go 1.22's new ServeMux with method-specific routing
	mux := http.NewServeMux()

	// Register API routes
	mux.HandleFunc("GET /api/healthz", gpuHandler.HealthCheck)
	mux.HandleFunc("GET /api/v1/gpu/metrics", gpuHandler.GetGPUMetrics)

	// Serve static files for frontend
	mux.Handle("GET /", http.FileServer(http.Dir("./static/")))

	// Apply middleware chain
	handler := middleware.Chain(
		mux,
		middleware.Logger,
		middleware.CORS,
		middleware.Recovery,
	)

	// Configure HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// getEnv retrieves environment variable value with fallback to default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
