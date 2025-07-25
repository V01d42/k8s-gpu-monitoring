package models

import "time"

// GPUMetrics represents GPU metrics data structure
type GPUMetrics struct {
	NodeName          string    `json:"node_name"`
	GPUIndex          int       `json:"gpu_index"`
	GPUName           string    `json:"gpu_name"`
	Utilization       float64   `json:"utilization"`
	MemoryUsed        float64   `json:"memory_used"`
	MemoryTotal       float64   `json:"memory_total"`
	MemoryFree        float64   `json:"memory_free"`
	MemoryUtilization float64   `json:"memory_utilization"`
	Temperature       float64   `json:"temperature"`
	PowerDraw         float64   `json:"power_draw"`
	PowerLimit        float64   `json:"power_limit"`
	Timestamp         time.Time `json:"timestamp"`
}

// GPUNode represents GPU node information
type GPUNode struct {
	NodeName  string   `json:"node_name"`
	GPUCount  int      `json:"gpu_count"`
	GPUModels []string `json:"gpu_models"`
}

// APIResponse represents standard API response structure
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// MetricsQuery represents Prometheus query parameters
type MetricsQuery struct {
	Query     string `json:"query"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Step      string `json:"step,omitempty"`
}
