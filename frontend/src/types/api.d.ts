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
  timestamp: string; // ISO8601
}
