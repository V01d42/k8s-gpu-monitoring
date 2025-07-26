export type GpuRow = {
  node_name: string;
  gpu_index: number;
  gpu_name: string;
  utilization: number;
  memory_used: number;
  memory_total: number;
  memory_utilization: number;
  temperature: number;
  power_draw: number;
  power_limit: number;
  timestamp: string;
};
