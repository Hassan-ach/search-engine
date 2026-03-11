import { collectDbMetrics } from "../db/inspector.js";
import { getCrawlQueueSize } from "../redis/inspector.js";
import { monitorStateGet, monitorStateSet } from "../db/queries.js";
import {
  ensureServiceUp,
  ensureServicesUp,
  runOneOffJob,
  countRunningJobsByImage,
} from "../docker/controller.js";
import { services } from "../docker/compose.js";
import { decideIndexer } from "../policy/indexer-policy.js";
import { decideSpider } from "../policy/spider-policy.js";
import { decideRanking } from "../policy/ranking-policy.js";
import { CleanupWorker } from "../workers/cleanup-worker.js";
import { EngineGuardian } from "../workers/engine-guardian.js";
import { logger } from "../logger/logger.js";
import { gauge, counter } from "../metrics/registry.js";
import { config } from "../config/env.js";

const STATE_LAST_INDEXER_SPAWN = "last_indexer_spawn_ms";
const STATE_LAST_SPIDER_SPAWN = "last_spider_spawn_ms";
const STATE_LAST_RANKING_INDEXED = "last_ranking_indexed_count";

export class Scheduler {
  private tickCount = 0;
  private timer: NodeJS.Timeout | null = null;
  private running = false;
  private readonly cleanup: CleanupWorker;
  private readonly guardian: EngineGuardian;

  constructor() {
    this.cleanup = new CleanupWorker(config.CLEANUP_BATCH_SIZE);
    this.guardian = new EngineGuardian();
  }

  start(): void {
    if (this.running) return;
    this.running = true;
    logger.info({ intervalMs: config.TICK_INTERVAL_MS }, "scheduler starting");
    void this._tick(); // immediate first tick
    this.timer = setInterval(() => void this._tick(), config.TICK_INTERVAL_MS);
  }

  stop(): void {
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
    this.running = false;
    logger.info("scheduler stopped");
  }

  private async _tick(): Promise<void> {
    this.tickCount++;
    const tickNo = this.tickCount;
    logger.debug({ tick: tickNo }, "tick start");
    gauge("scheduler_tick_count", tickNo, "current scheduler tick number");

    try {
      await this._ensureBaseline();
      await this._tickIndexer();
      await this._tickSpider();
      await this._tickRanking();

      if (tickNo % config.CLEANUP_EVERY_N_TICKS === 0) {
        await this.cleanup.run();
      }

      await this.guardian.check();
    } catch (err) {
      logger.error({ err, tick: tickNo }, "tick error");
      counter("scheduler_tick_errors_total", "total tick errors");
    }

    logger.debug({ tick: tickNo }, "tick done");
  }

  private async _ensureBaseline(): Promise<void> {
    await ensureServicesUp(services.psql.compose, [
      services.psql.service,
      services.redis.service,
    ]);
    if (config.MANAGE_ADMINER) {
      await ensureServiceUp(services.adminer.compose, services.adminer.service).catch(
        (err) => logger.warn({ err }, "adminer ensure failed (non-fatal)")
      );
    }
  }

  private async _tickIndexer(): Promise<void> {
    const [dbMetrics, runningJobs, rawLastSpawn] = await Promise.all([
      collectDbMetrics(),
      countRunningJobsByImage(
        services.indexer.compose,
        services.indexer.service
      ),
      monitorStateGet(STATE_LAST_INDEXER_SPAWN),
    ]);

    gauge("indexer_unindexed_backlog", dbMetrics.unindexedPages, "unindexed page count");
    gauge("indexer_running_jobs", runningJobs, "active indexer containers");

    const decision = decideIndexer({
      unindexedPages: dbMetrics.unindexedPages,
      runningIndexerJobs: runningJobs,
      lastSpawnedAt: rawLastSpawn !== null ? Number(rawLastSpawn) : null,
      nowMs: Date.now(),
      threshold: config.INDEXER_BACKLOG_THRESHOLD,
      maxParallel: config.INDEXER_MAX_PARALLEL,
      cooldownMs: config.INDEXER_COOLDOWN_MS,
    });

    logger.debug({ decision }, "indexer policy");

    if (decision.shouldSpawn) {
      await runOneOffJob(services.indexer.compose, services.indexer.service);
      await monitorStateSet(STATE_LAST_INDEXER_SPAWN, BigInt(Date.now()));
      counter("indexer_jobs_spawned_total", "total indexer jobs spawned");
      logger.info({ reason: decision.reason }, "indexer job spawned");
    }
  }

  private async _tickSpider(): Promise<void> {
    const [dbMetrics, crawlQueueSize, runningSpiders, rawLastSpawn] =
      await Promise.all([
        collectDbMetrics(),
        getCrawlQueueSize(),
        countRunningJobsByImage(services.spider.compose, services.spider.service),
        monitorStateGet(STATE_LAST_SPIDER_SPAWN),
      ]);

    gauge("spider_queue_size", crawlQueueSize, "redis crawl queue size");
    gauge("spider_crawl_rate_per_min", dbMetrics.pagesCrawledLastMinute, "pages crawled last minute");
    gauge("spider_running_count", runningSpiders, "running spider containers");

    const decision = decideSpider({
      crawlQueueSize,
      crawlRatePerMinute: dbMetrics.pagesCrawledLastMinute,
      runningSpiders,
      lastSpawnedAt: rawLastSpawn !== null ? Number(rawLastSpawn) : null,
      nowMs: Date.now(),
      minInstances: config.SPIDER_MIN_INSTANCES,
      maxInstances: config.SPIDER_MAX_INSTANCES,
      rateLowThreshold: config.SPIDER_RATE_LOW_THRESHOLD,
      queueHighThreshold: config.SPIDER_QUEUE_HIGH_THRESHOLD,
      cooldownMs: config.SPIDER_COOLDOWN_MS,
    });

    logger.debug({ decision }, "spider policy");

    if (decision.shouldSpawn) {
      await ensureServiceUp(services.spider.compose, services.spider.service);
      await monitorStateSet(STATE_LAST_SPIDER_SPAWN, BigInt(Date.now()));
      counter("spider_instances_started_total", "total spider instances started");
      logger.info({ reason: decision.reason }, "spider instance started");
    }
  }

  private async _tickRanking(): Promise<void> {
    const [dbMetrics, runningJobs, lastRankingIndexed] = await Promise.all([
      collectDbMetrics(),
      countRunningJobsByImage(services.ranking.compose, services.ranking.service),
      monitorStateGet(STATE_LAST_RANKING_INDEXED),
    ]);

    gauge("ranking_running_jobs", runningJobs, "active ranking containers");
    gauge("ranking_total_indexed", dbMetrics.totalIndexedPages, "total indexed pages");

    const decision = decideRanking({
      totalIndexedPages: dbMetrics.totalIndexedPages,
      lastRankingAt: lastRankingIndexed,
      triggerDelta: config.RANKING_TRIGGER_DELTA,
      runningRankingJobs: runningJobs,
      maxParallel: config.RANKING_MAX_PARALLEL,
    });

    logger.debug({ decision }, "ranking policy");

    if (decision.shouldRun) {
      await runOneOffJob(services.ranking.compose, services.ranking.service);
      await monitorStateSet(
        STATE_LAST_RANKING_INDEXED,
        BigInt(dbMetrics.totalIndexedPages)
      );
      counter("ranking_jobs_triggered_total", "total ranking jobs triggered");
      logger.info({ reason: decision.reason }, "ranking job triggered");
    }
  }
}
