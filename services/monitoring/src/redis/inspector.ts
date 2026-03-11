import { redisClient } from "./client.js";
import { logger } from "../logger/logger.js";

const CRAWL_QUEUE_KEY = "urls";

export async function getCrawlQueueSize(): Promise<number> {
  try {
    const size = await redisClient.zCard(CRAWL_QUEUE_KEY);
    return size;
  } catch (err) {
    logger.error({ err }, "failed to get crawl queue size");
    throw err;
  }
}
