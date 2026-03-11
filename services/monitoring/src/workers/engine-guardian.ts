import http from "node:http";
import { ensureServiceUp, isContainerRunning } from "../docker/controller.js";
import { services } from "../docker/compose.js";
import { logger } from "../logger/logger.js";
import { gauge, counter } from "../metrics/registry.js";
import { config } from "../config/env.js";

interface RestartWindow {
  timestamps: number[];
  windowMs: number;
  maxRestarts: number;
}

function pruneWindow(w: RestartWindow, nowMs: number): void {
  w.timestamps = w.timestamps.filter((t) => nowMs - t < w.windowMs);
}

export class EngineGuardian {
  private restartWindow: RestartWindow = {
    timestamps: [],
    windowMs: 5 * 60 * 1000, // 5 minutes
    maxRestarts: config.ENGINE_MAX_RESTARTS_PER_5M,
  };

  private degraded = false;

  async check(): Promise<void> {
    const running = await isContainerRunning(config.ENGINE_CONTAINER_NAME);
    gauge("engine_running", running ? 1 : 0, "1 if engine container is running");

    if (!running) {
      await this._handleDown();
      return;
    }

    // Optional HTTP liveness probe
    const alive = await this._httpProbe(config.ENGINE_HTTP_CHECK_URL);
    gauge("engine_http_alive", alive ? 1 : 0, "1 if engine HTTP probe succeeded");
    if (!alive) {
      logger.warn("engine HTTP probe failed (container running but not responding)");
    }
    if (this.degraded && running && alive) {
      this.degraded = false;
      logger.info("engine recovered");
    }
  }

  private async _handleDown(): Promise<void> {
    const nowMs = Date.now();
    pruneWindow(this.restartWindow, nowMs);

    if (this.restartWindow.timestamps.length >= this.restartWindow.maxRestarts) {
      if (!this.degraded) {
        logger.error(
          { restarts: this.restartWindow.timestamps.length },
          "engine restart rate too high — entering degraded mode"
        );
        this.degraded = true;
      }
      counter("engine_restarts_rate_limited", "times engine restart was rate-limited");
      return;
    }

    logger.warn("engine is down — attempting restart");
    try {
      await ensureServiceUp(services.engine.compose, services.engine.service);
      this.restartWindow.timestamps.push(nowMs);
      counter("engine_restarts_total", "total engine restart attempts");
      logger.info("engine restart triggered");
    } catch (err) {
      logger.error({ err }, "engine restart failed");
      counter("engine_restart_failures_total", "total engine restart failures");
    }
  }

  private _httpProbe(url: string): Promise<boolean> {
    return new Promise((resolve) => {
      const timeout = setTimeout(() => resolve(false), 3000);
      http
        .get(url, (res) => {
          clearTimeout(timeout);
          resolve((res.statusCode ?? 500) < 500);
          res.resume();
        })
        .on("error", () => {
          clearTimeout(timeout);
          resolve(false);
        });
    });
  }
}
