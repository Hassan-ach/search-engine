import { logger } from "../logger/logger.js";

export interface RetryOptions {
  maxAttempts: number;
  baseDelayMs: number;
  label: string;
}

function jitter(ms: number): number {
  return ms + Math.floor(Math.random() * ms * 0.25);
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export async function withRetry<T>(
  fn: () => Promise<T>,
  opts: RetryOptions
): Promise<T> {
  let lastError: unknown;
  for (let attempt = 1; attempt <= opts.maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (err) {
      lastError = err;
      if (attempt < opts.maxAttempts) {
        const backoff = jitter(opts.baseDelayMs * 2 ** (attempt - 1));
        logger.warn(
          { label: opts.label, attempt, backoff, err },
          "retrying after error"
        );
        await delay(backoff);
      }
    }
  }
  logger.error(
    { label: opts.label, maxAttempts: opts.maxAttempts, err: lastError },
    "all retry attempts exhausted"
  );
  throw lastError;
}
