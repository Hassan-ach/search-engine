import { z } from "zod";

const boolStr = z
  .string()
  .transform((v) => v.toLowerCase() === "true")
  .or(z.boolean());

export const configSchema = z.object({
  // Monitor HTTP server
  MONITOR_PORT: z.coerce.number().int().positive().default(7070),

  // Logging
  LOG_PATH: z.string().default("logs/monitor.log"),

  // Scheduler
  TICK_INTERVAL_MS: z.coerce.number().int().positive().default(15000),

  // Compose file paths (relative to project root)
  ROOT_COMPOSE_FILE: z.string().default("docker-compose.yml"),
  ENGINE_COMPOSE_FILE: z
    .string()
    .default("services/engine/docker-compose.yml"),
  SPIDER_COMPOSE_FILE: z
    .string()
    .default("services/spider/docker-compose.yml"),
  INDEXER_COMPOSE_FILE: z
    .string()
    .default("services/indexer/docker-compose.yml"),
  RANKING_COMPOSE_FILE: z
    .string()
    .default("services/ranking/docker-compose.yml"),

  // Data sources
  DATABASE_URL: z.string().url(),
  REDIS_URL: z.string(),

  // Engine
  ENGINE_CONTAINER_NAME: z.string().default("engine"),
  ENGINE_HTTP_CHECK_URL: z.string().url().default("http://localhost:1323/"),
  ENGINE_RESTART_BACKOFF_MS: z.coerce.number().int().positive().default(5000),
  ENGINE_MAX_RESTARTS_PER_5M: z.coerce.number().int().positive().default(5),

  // Indexer
  INDEXER_BACKLOG_THRESHOLD: z.coerce.number().int().positive().default(100),
  INDEXER_MAX_PARALLEL: z.coerce.number().int().positive().default(3),
  INDEXER_COOLDOWN_MS: z.coerce.number().int().positive().default(20000),

  // Spider
  SPIDER_MIN_INSTANCES: z.coerce.number().int().positive().default(1),
  SPIDER_MAX_INSTANCES: z.coerce.number().int().positive().default(4),
  SPIDER_RATE_LOW_THRESHOLD: z.coerce.number().int().min(0).default(20),
  SPIDER_QUEUE_HIGH_THRESHOLD: z.coerce.number().int().positive().default(500),
  SPIDER_COOLDOWN_MS: z.coerce.number().int().positive().default(30000),

  // Ranking
  RANKING_TRIGGER_DELTA: z.coerce.number().int().positive().default(50),
  RANKING_MAX_PARALLEL: z.coerce.number().int().positive().default(1),

  // Cleanup
  CLEANUP_BATCH_SIZE: z.coerce.number().int().positive().default(500),
  CLEANUP_EVERY_N_TICKS: z.coerce.number().int().positive().default(4),

  // Optional adminer management
  MANAGE_ADMINER: boolStr.default("false"),
});

export type Config = z.infer<typeof configSchema>;
