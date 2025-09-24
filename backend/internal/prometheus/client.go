package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s-gpu-monitoring/internal/models"
	"k8s-gpu-monitoring/internal/timeutil"
)

// Client represents a Prometheus HTTP API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// PrometheusResponse represents the response structure from Prometheus API.
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

// NewClient creates a new Prometheus client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Query executes a PromQL query.
func (c *Client) Query(ctx context.Context, query string) (*PrometheusResponse, error) {
	queryURL := fmt.Sprintf("%s/api/v1/query", c.baseURL)

	params := url.Values{}
	params.Add("query", query)
	params.Add("time", strconv.FormatInt(time.Now().Unix(), 10))

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var promResp PrometheusResponse
	if err := json.Unmarshal(body, &promResp); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	if promResp.Status != "success" {
		return nil, fmt.Errorf("prometheus query failed: %s - %s", promResp.ErrorType, promResp.Error)
	}

	return &promResp, nil
}

// GetGPUMetrics retrieves GPU metrics from Prometheus with concurrent queries.
func (c *Client) GetGPUMetrics(ctx context.Context) ([]models.GPUMetrics, error) {
	// Execute multiple queries concurrently
	queries := map[string]string{
		"utilization":        `nvidia_gpu_utilization_percent`,
		"memory_used":        `nvidia_gpu_used_memory_bytes`,
		"memory_total":       `nvidia_gpu_total_memory_bytes`,
		"memory_free":        `nvidia_gpu_free_memory_bytes`,
		"memory_utilization": `nvidia_gpu_memory_utilization_percent`,
		"temperature":        `nvidia_gpu_temperature_celsius`,
	}

	results := make(map[string]*PrometheusResponse)
	errors := make(chan error, len(queries))
	var mu sync.Mutex

	// Execute queries concurrently
	for name, query := range queries {
		go func(name, query string) {
			resp, err := c.Query(ctx, query)
			if err != nil {
				errors <- fmt.Errorf("query %s failed: %w", name, err)
				return
			}
			mu.Lock()
			results[name] = resp
			mu.Unlock()
			errors <- nil
		}(name, query)
	}

	// Wait for all queries to complete
	for i := 0; i < len(queries); i++ {
		if err := <-errors; err != nil {
			return nil, err
		}
	}

	return c.parseGPUMetrics(results)
}

// GetGPUProcesses retrieves running GPU processes from Prometheus.
func (c *Client) GetGPUProcesses(ctx context.Context) ([]models.GPUProcess, error) {
	queries := map[string]string{
		"gpu_memory": `nvidia_gpu_process_gpu_memory_bytes`,
		"cpu_usage":  `nvidia_gpu_process_cpu_percent`,
		"mem_usage":  `nvidia_gpu_process_memory_percent`,
	}

	results := make(map[string]*PrometheusResponse)
	errors := make(chan error, len(queries))
	var mu sync.Mutex

	for name, query := range queries {
		go func(name, query string) {
			resp, err := c.Query(ctx, query)
			if err != nil {
				errors <- fmt.Errorf("query %s failed: %w", name, err)
				return
			}
			mu.Lock()
			results[name] = resp
			mu.Unlock()
			errors <- nil
		}(name, query)
	}

	for i := 0; i < len(queries); i++ {
		if err := <-errors; err != nil {
			return nil, err
		}
	}

	return c.parseGPUProcesses(results)
}

// parseGPUMetrics parses Prometheus response into GPUMetrics.
func (c *Client) parseGPUMetrics(results map[string]*PrometheusResponse) ([]models.GPUMetrics, error) {
	// Group metrics by node and GPU index
	metricsMap := make(map[string]*models.GPUMetrics) // key: "node_name:gpu_index"

	for metricType, response := range results {
		for _, result := range response.Data.Result {
			nodeName := result.Metric["hostname"]
			gpuIndex := result.Metric["gpu_id"]
			gpuName := result.Metric["gpu_name"]

			if nodeName == "" || gpuIndex == "" {
				continue
			}

			key := fmt.Sprintf("%s:%s", nodeName, gpuIndex)

			if metricsMap[key] == nil {
				idx, _ := strconv.Atoi(gpuIndex)
				metricsMap[key] = &models.GPUMetrics{
					NodeName:  nodeName,
					GPUIndex:  idx,
					GPUName:   gpuName,
					Timestamp: timeutil.NowJST(),
				}
			}

			// Parse and extract value
			if len(result.Value) >= 2 {
				valueStr, ok := result.Value[1].(string)
				if !ok {
					continue
				}

				value, err := strconv.ParseFloat(valueStr, 64)
				if err != nil {
					continue
				}

				// Set value based on metric type
				switch metricType {
				case "utilization":
					metricsMap[key].Utilization = value
				case "memory_used":
					metricsMap[key].MemoryUsed = int(value)
				case "memory_total":
					metricsMap[key].MemoryTotal = int(value)
				case "memory_free":
					metricsMap[key].MemoryFree = int(value)
				case "memory_utilization":
					metricsMap[key].MemoryUtilization = int(value)
				case "temperature":
					metricsMap[key].Temperature = int(value)
				}
			}
		}
	}

	// Convert to slice
	var gpuMetrics []models.GPUMetrics
	for _, metrics := range metricsMap {
		gpuMetrics = append(gpuMetrics, *metrics)
	}

	return gpuMetrics, nil
}

// parseGPUProcesses parses Prometheus response into GPUProcess slice.
func (c *Client) parseGPUProcesses(results map[string]*PrometheusResponse) ([]models.GPUProcess, error) {
	processMap := make(map[string]*models.GPUProcess)

	for metricType, response := range results {
		if response == nil {
			continue
		}

		for _, result := range response.Data.Result {
			nodeName := result.Metric["hostname"]
			gpuIndex := result.Metric["gpu_id"]
			pidStr := result.Metric["pid"]

			if nodeName == "" || gpuIndex == "" || pidStr == "" {
				continue
			}

			key := fmt.Sprintf("%s:%s:%s", nodeName, gpuIndex, pidStr)

			if processMap[key] == nil {
				idx, _ := strconv.Atoi(gpuIndex)
				pid, _ := strconv.Atoi(pidStr)

				processMap[key] = &models.GPUProcess{
					NodeName:    nodeName,
					GPUIndex:    idx,
					PID:         pid,
					ProcessName: result.Metric["process_name"],
					User:        result.Metric["user"],
					Command:     result.Metric["command"],
					Timestamp:   timeutil.NowJST(),
				}
			}

			if len(result.Value) < 2 {
				continue
			}

			valueStr, ok := result.Value[1].(string)
			if !ok {
				continue
			}

			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				continue
			}

			switch metricType {
			case "gpu_memory":
				processMap[key].GPUMemory = int(value)
			case "cpu_usage":
				processMap[key].CPU = int(value)
			case "mem_usage":
				processMap[key].Memory = int(value)
			}
		}
	}

	if len(processMap) == 0 {
		return []models.GPUProcess{}, nil
	}

	processes := make([]models.GPUProcess, 0, len(processMap))
	for _, proc := range processMap {
		processes = append(processes, *proc)
	}

	sort.Slice(processes, func(i, j int) bool {
		if processes[i].NodeName != processes[j].NodeName {
			return processes[i].NodeName < processes[j].NodeName
		}
		if processes[i].GPUIndex != processes[j].GPUIndex {
			return processes[i].GPUIndex < processes[j].GPUIndex
		}
		return processes[i].PID < processes[j].PID
	})

	return processes, nil
}
