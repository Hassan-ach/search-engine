import {
  countUnindexedPages,
  countIndexedPagesTotal,
  countPagesCrawledLastMinute,
} from "./queries.js";
import { logger } from "../logger/logger.js";

export interface DbMetrics {
  unindexedPages: number;
  totalIndexedPages: number;
  pagesCrawledLastMinute: number;
}

export async function collectDbMetrics(): Promise<DbMetrics> {
  try {
    const [unindexedPages, totalIndexedPages, pagesCrawledLastMinute] =
      await Promise.all([
        countUnindexedPages(),
        countIndexedPagesTotal(),
        countPagesCrawledLastMinute(),
      ]);
    return { unindexedPages, totalIndexedPages, pagesCrawledLastMinute };
  } catch (err) {
    logger.error({ err }, "failed to collect db metrics");
    throw err;
  }
}
