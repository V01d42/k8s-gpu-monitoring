package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"k8s-gpu-monitoring/internal/models"
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

	// Execute queries concurrently
	for name, query := range queries {
		go func(name, query string) {
			resp, err := c.Query(ctx, query)
			if err != nil {
				errors <- fmt.Errorf("query %s failed: %w", name, err)
				return
			}
			results[name] = resp
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
					Timestamp: time.Now(),
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
					metricsMap[key].Utilization = value // Already in percentage format
				case "memory_used":
					metricsMap[key].MemoryUsed = value / (1024 * 1024 * 1024) // bytes to GB
				case "memory_total":
					metricsMap[key].MemoryTotal = value / (1024 * 1024 * 1024) // bytes to GB
				case "memory_free":
					metricsMap[key].MemoryFree = value / (1024 * 1024 * 1024) // bytes to GB
				case "memory_utilization":
					metricsMap[key].MemoryUtilization = value // Already in percentage format
				case "temperature":
					metricsMap[key].Temperature = value
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

// GetGPUNodes retrieves GPU node information.
func (c *Client) GetGPUNodes(ctx context.Context) ([]models.GPUNode, error) {
	query := `group by (hostname, gpu_name) (nvidia_gpu_utilization_percent)`

	resp, err := c.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("getting GPU nodes: %w", err)
	}

	var nodes []models.GPUNode
	nodeMap := make(map[string]*models.GPUNode)

	for _, result := range resp.Data.Result {
		nodeName := result.Metric["hostname"]
		gpuName := result.Metric["gpu_name"]
		if nodeName == "" {
			continue
		}

		if nodeMap[nodeName] == nil {
			nodeMap[nodeName] = &models.GPUNode{
				NodeName:  nodeName,
				GPUModels: make([]string, 0),
			}
		}

		// Add GPU model if not already present
		if gpuName != "" {
			found := false
			for _, existingModel := range nodeMap[nodeName].GPUModels {
				if existingModel == gpuName {
					found = true
					break
				}
			}
			if !found {
				nodeMap[nodeName].GPUModels = append(nodeMap[nodeName].GPUModels, gpuName)
			}
		}

		// Separate query to get GPU count
		countQuery := fmt.Sprintf(`count by (hostname) (nvidia_gpu_utilization_percent{hostname="%s"})`, nodeName)
		countResp, err := c.Query(ctx, countQuery)
		if err == nil && len(countResp.Data.Result) > 0 && len(countResp.Data.Result[0].Value) >= 2 {
			if countStr, ok := countResp.Data.Result[0].Value[1].(string); ok {
				if count, err := strconv.Atoi(countStr); err == nil {
					nodeMap[nodeName].GPUCount = count
				}
			}
		}
	}

	for _, node := range nodeMap {
		nodes = append(nodes, *node)
	}

	return nodes, nil
}
