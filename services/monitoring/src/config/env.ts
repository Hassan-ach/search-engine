import path from "node:path";
import fs from "node:fs";
import { configSchema, type Config } from "./schema.js";

function loadDotEnv(filePath: string): void {
  if (!fs.existsSync(filePath)) return;
  const lines = fs.readFileSync(filePath, "utf-8").split("\n");
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const eqIdx = trimmed.indexOf("=");
    if (eqIdx === -1) continue;
    const key = trimmed.slice(0, eqIdx).trim();
    const val = trimmed.slice(eqIdx + 1).trim();
    if (!(key in process.env)) {
      process.env[key] = val;
    }
  }
}

// Resolve the monitoring service root (this file lives at src/config/)
export const MONITORING_ROOT = path.resolve(
  path.dirname(new URL(import.meta.url).pathname),
  "../.."
);

// Project root is two levels up from services/monitoring
export const PROJECT_ROOT = path.resolve(MONITORING_ROOT, "../..");

// Load .env if present
loadDotEnv(path.join(MONITORING_ROOT, ".env"));

const parseResult = configSchema.safeParse(process.env);
if (!parseResult.success) {
  console.error("❌ Invalid configuration:");
  for (const issue of parseResult.error.issues) {
    console.error(`  ${issue.path.join(".")}: ${issue.message}`);
  }
  process.exit(1);
}

export const config: Config & {
  MONITORING_ROOT: string;
  PROJECT_ROOT: string;
} = {
  ...parseResult.data,
  MONITORING_ROOT,
  PROJECT_ROOT,
};
