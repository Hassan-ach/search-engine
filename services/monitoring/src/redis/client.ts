import { createClient } from "redis";
import { config } from "../config/env.js";
import { logger } from "../logger/logger.js";

export const redisClient = createClient({ url: config.REDIS_URL });

redisClient.on("error", (err) => {
  logger.error({ err }, "redis client error");
});

let connected = false;

export async function connectRedis(): Promise<void> {
  if (!connected) {
    // Create a fresh client each attempt so it's not in an error state
    if (redisClient.isOpen) {
      connected = true;
      return;
    }
    await redisClient.connect();
    connected = true;
    logger.info("redis connected");
  }
}

export async function disconnectRedis(): Promise<void> {
  if (connected) {
    await redisClient.disconnect();
    connected = false;
  }
}
