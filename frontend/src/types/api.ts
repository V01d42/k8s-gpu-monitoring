// API response types for backend endpoints

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface GPUMetrics {
  node_name: string;
  gpu_index: number;
  gpu_name: string;
  utilization: number;
  memory_used: number;
  memory_total: number;
  memory_free: number;
  memory_utilization: number;
  temperature: number;
  power_draw: number;
  power_limit: number;
  timestamp: string; // ISO8601
}

export interface GPUNode {
  node_name: string;
  gpu_count: number;
  gpu_models: string[];
}

// For /api/v1/gpu/utilization (simplified)
export interface GPUUtilization {
  node: string;
  gpu_index: string | number;
  utilization: string | number;
  timestamp: string | number;
}
