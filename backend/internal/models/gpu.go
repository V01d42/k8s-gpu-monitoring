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
	MemoryUtilization float64 `json:"memory_utilization"`
	Temperature       int     `json:"temperature"`
	Timestamp         string  `json:"timestamp"`
}

// GPUProcess represents running GPU-related processes and their usage metrics.
type GPUProcess struct {
	NodeName    string  `json:"node_name"`
	GPUIndex    int     `json:"gpu_index"`
	PID         int     `json:"pid"`
	ProcessName string  `json:"process_name"`
	User        string  `json:"user"`
	Command     string  `json:"command"`
	GPUMemory   int     `json:"gpu_memory"`
	CPU         float64 `json:"cpu"`
	Memory      float64 `json:"memory"`
	Timestamp   string  `json:"timestamp"`
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
