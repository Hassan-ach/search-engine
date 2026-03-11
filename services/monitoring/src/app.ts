import express from "express";
import { config } from "./config/env.js";
import { logger } from "./logger/logger.js";
import { metricsRouter } from "./metrics/http.js";
import { Scheduler } from "./scheduler/scheduler.js";
import { pool, query } from "./db/client.js";
import { ensureMonitorStateTable } from "./db/queries.js";
import { connectRedis, disconnectRedis } from "./redis/client.js";
import { ensureServicesUp, ensureServiceUp } from "./docker/controller.js";
import { services } from "./docker/compose.js";

async function waitForDb(maxAttempts = 20, intervalMs = 3000): Promise<void> {
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      await query("SELECT 1");
      logger.info("database ready");
      return;
    } catch (err) {
      logger.info({ attempt, maxAttempts, err }, "waiting for database...");
      await new Promise((r) => setTimeout(r, intervalMs));
    }
  }
  throw new Error("database did not become ready in time");
}

async function waitForRedis(maxAttempts = 20, intervalMs = 3000): Promise<void> {
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      await connectRedis();
      return;
    } catch {
      logger.info({ attempt, maxAttempts }, "waiting for redis...");
      await new Promise((r) => setTimeout(r, intervalMs));
    }
  }
  throw new Error("redis did not become ready in time");
}

export async function createApp(): Promise<{
  server: ReturnType<typeof express.application.listen>;
  scheduler: Scheduler;
  shutdown: () => Promise<void>;
}> {
  // Step 1: Boot baseline infra services via Docker before connecting
  // Start psql + redis with one compose call to avoid sequential mutex waits
  logger.info("starting baseline infrastructure services...");
  await ensureServicesUp(services.psql.compose, [
    services.psql.service,
    services.redis.service,
  ]);
  if (config.MANAGE_ADMINER) {
    await ensureServiceUp(services.adminer.compose, services.adminer.service).catch(
      (err) => logger.warn({ err }, "adminer ensure failed (non-fatal)")
    );
  }
  logger.info("baseline services requested");

  // Step 2: Wait for DB and Redis to accept connections
  await waitForDb();
  await waitForRedis();

  // Step 3: Initialize DB schema additions
  await ensureMonitorStateTable();
  logger.info("monitor_state table ensured");

  const app = express();
  app.use(express.json());
  app.use(metricsRouter);

  const scheduler = new Scheduler();

  const server = app.listen(config.MONITOR_PORT, () => {
    logger.info({ port: config.MONITOR_PORT }, "HTTP server listening");
  });

  scheduler.start();

  async function shutdown(): Promise<void> {
    logger.info("shutting down monitor");
    scheduler.stop();
    server.close();
    await pool.end();
    await disconnectRedis();
    logger.info("monitor shutdown complete");
  }

  return { server, scheduler, shutdown };
}
