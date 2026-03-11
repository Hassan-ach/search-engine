import { createApp } from "./app.js";
import { logger } from "./logger/logger.js";

async function main(): Promise<void> {
  logger.info("boogle monitor starting");

  const { shutdown } = await createApp();

  const onSignal = async (signal: string) => {
    logger.info({ signal }, "received signal");
    try {
      await shutdown();
    } finally {
      process.exit(0);
    }
  };

  process.on("SIGINT", () => void onSignal("SIGINT"));
  process.on("SIGTERM", () => void onSignal("SIGTERM"));
}

main().catch((err) => {
  console.error("fatal startup error:", err);
  process.exit(1);
});
