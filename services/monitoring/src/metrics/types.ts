export type MetricType = "counter" | "gauge";

export interface MetricEntry {
  type: MetricType;
  value: number;
  description: string;
  labels?: Record<string, string>;
  updatedAt: string;
}

export type MetricsSnapshot = Record<string, MetricEntry>;
