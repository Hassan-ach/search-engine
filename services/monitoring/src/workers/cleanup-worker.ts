import { cleanupIndexedHtml } from "../db/queries.js";
import { logger } from "../logger/logger.js";
import { gauge, counterInc } from "../metrics/registry.js";

export class CleanupWorker {
  private readonly batchSize: number;

  constructor(batchSize: number) {
    this.batchSize = batchSize;
  }

  async run(): Promise<void> {
    try {
      const cleaned = await cleanupIndexedHtml(this.batchSize);
      gauge("cleanup_last_batch_rows", cleaned, "rows cleared in last cleanup batch");
      counterInc("cleanup_rows_total", "total html rows cleared", cleaned);
      logger.info({ cleaned }, "cleanup batch complete");
    } catch (err) {
      logger.error({ err }, "cleanup worker error");
    }
  }
}
