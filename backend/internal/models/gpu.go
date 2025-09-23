package models

// GPUMetrics represents GPU metrics data structure
type GPUMetrics struct {
	NodeName          string  `json:"node_name"`
	GPUIndex          int     `json:"gpu_index"`
	GPUName           string  `json:"gpu_name"`
	Utilization       float64 `json:"utilization"`
	MemoryUsed        int     `json:"memory_used"`
	MemoryTotal       int     `json:"memory_total"`
	MemoryFree        int     `json:"memory_free"`
	MemoryUtilization int     `json:"memory_utilization"`
	Temperature       int     `json:"temperature"`
	Timestamp         string  `json:"timestamp"`
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
