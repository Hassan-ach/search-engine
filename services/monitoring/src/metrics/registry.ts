import type { MetricEntry, MetricsSnapshot } from "./types.js";

const store = new Map<string, MetricEntry>();
const startTime = Date.now();

function now(): string {
  return new Date().toISOString();
}

export function counter(
  name: string,
  description: string,
  labels?: Record<string, string>
): void {
  const existing = store.get(name);
  store.set(name, {
    type: "counter",
    value: existing ? existing.value + 1 : 1,
    description,
    labels,
    updatedAt: now(),
  });
}

export function counterInc(
  name: string,
  description: string,
  by: number,
  labels?: Record<string, string>
): void {
  const existing = store.get(name);
  store.set(name, {
    type: "counter",
    value: existing ? existing.value + by : by,
    description,
    labels,
    updatedAt: now(),
  });
}

export function gauge(
  name: string,
  value: number,
  description: string,
  labels?: Record<string, string>
): void {
  store.set(name, { type: "gauge", value, description, labels, updatedAt: now() });
}

export interface FullSnapshot {
  uptime_seconds: number;
  metrics: MetricsSnapshot;
}

export function getSnapshot(): FullSnapshot {
  const metrics: MetricsSnapshot = {};
  for (const [key, entry] of store.entries()) {
    metrics[key] = entry;
  }
  return {
    uptime_seconds: Math.floor((Date.now() - startTime) / 1000),
    metrics,
  };
}
