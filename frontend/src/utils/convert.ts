import type { GPUMetrics } from "../types/api";

const convertGBtoMiB = (bytes: number) => Math.round(bytes / 2 ** 20);

export const convertGPUMetrics = (ms: GPUMetrics[]) => {
  const convertKeys = ["memory_used", "memory_total", "memory_free"] as const;
  for (const m of ms) {
    for (const key of convertKeys) {
      m[key] = convertGBtoMiB(m[key]);
    }
  }
};
