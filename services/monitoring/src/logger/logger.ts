import fs from "node:fs";
import path from "node:path";
import pino from "pino";
import { config } from "../config/env.js";

const logPath = path.isAbsolute(config.LOG_PATH)
  ? config.LOG_PATH
  : path.join(config.MONITORING_ROOT, config.LOG_PATH);

// Ensure log directory exists
fs.mkdirSync(path.dirname(logPath), { recursive: true });

const fileTransport = pino.transport({
  targets: [
    {
      target: "pino/file",
      options: { destination: logPath, mkdir: true },
      level: "debug",
    },
    {
      target: "pino-pretty",
      options: { colorize: true, translateTime: "SYS:standard" },
      level: "info",
    },
  ],
});

export const logger = pino(
  {
    level: "debug",
    timestamp: pino.stdTimeFunctions.isoTime,
  },
  fileTransport
);
