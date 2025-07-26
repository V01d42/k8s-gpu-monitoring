// Mock API responses for backend endpoints
import type { ApiResponse, GPUMetrics, GPUNode, GPUUtilization } from "./api";

export const mockGpuMetrics: ApiResponse<GPUMetrics[]> = {
  success: true,
  message: "GPU metrics retrieved successfully",
  data: [
    {
      node_name: "gpu14",
      gpu_index: 0,
      gpu_name: "NVIDIA A100",
      utilization: 85.2,
      memory_used: 32,
      memory_total: 40,
      memory_free: 8,
      memory_utilization: 80.0,
      temperature: 70.5,
      power_draw: 250.0,
      power_limit: 300.0,
      timestamp: "2025-07-26T12:00:00Z",
    },
    {
      node_name: "gpu15",
      gpu_index: 1,
      gpu_name: "NVIDIA V100",
      utilization: 60.1,
      memory_used: 12,
      memory_total: 16,
      memory_free: 4,
      memory_utilization: 75.0,
      temperature: 65.0,
      power_draw: 180.0,
      power_limit: 250.0,
      timestamp: "2025-07-26T12:00:00Z",
    },
  ],
};

export const mockGpuNodes: ApiResponse<GPUNode[]> = {
  success: true,
  message: "GPU nodes retrieved successfully",
  data: [
    {
      node_name: "gpu14",
      gpu_count: 2,
      gpu_models: ["NVIDIA A100"],
    },
    {
      node_name: "gpu15",
      gpu_count: 1,
      gpu_models: ["NVIDIA V100"],
    },
  ],
};

export const mockGpuUtilization: ApiResponse<GPUUtilization[]> = {
  success: true,
  message: "GPU utilization retrieved successfully",
  data: [
    {
      node: "gpu14",
      gpu_index: 0,
      utilization: 85.2,
      timestamp: 1721995200,
    },
    {
      node: "gpu15",
      gpu_index: 1,
      utilization: 60.1,
      timestamp: 1721995200,
    },
  ],
};

export const mockHealthz: ApiResponse = {
  success: true,
  message: "Service is healthy",
  data: {
    status: "healthy",
    timestamp: "2025-07-26T12:00:00Z",
    version: "1.0.0",
  },
};
