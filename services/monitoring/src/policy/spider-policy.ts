export interface SpiderPolicyInput {
  crawlQueueSize: number;
  crawlRatePerMinute: number;
  runningSpiders: number;
  lastSpawnedAt: number | null; // epoch ms
  nowMs: number;
  minInstances: number;
  maxInstances: number;
  rateLowThreshold: number;
  queueHighThreshold: number;
  cooldownMs: number;
}

export interface SpiderPolicyDecision {
  shouldSpawn: boolean;
  reason: string;
}

export function decideSpider(input: SpiderPolicyInput): SpiderPolicyDecision {
  // Always ensure at least minInstances
  if (input.runningSpiders < input.minInstances) {
    return {
      shouldSpawn: true,
      reason: `below min instances (${input.runningSpiders} < ${input.minInstances})`,
    };
  }
  if (input.runningSpiders >= input.maxInstances) {
    return {
      shouldSpawn: false,
      reason: `already at max instances ${input.maxInstances}`,
    };
  }
  if (input.lastSpawnedAt !== null) {
    const elapsed = input.nowMs - input.lastSpawnedAt;
    if (elapsed < input.cooldownMs) {
      return {
        shouldSpawn: false,
        reason: `cooldown active (${Math.round(elapsed / 1000)}s / ${Math.round(input.cooldownMs / 1000)}s)`,
      };
    }
  }
  const queuePressure = input.crawlQueueSize >= input.queueHighThreshold;
  const rateLow = input.crawlRatePerMinute <= input.rateLowThreshold;
  if (queuePressure && rateLow) {
    return {
      shouldSpawn: true,
      reason: `queue ${input.crawlQueueSize} >= ${input.queueHighThreshold} and rate ${input.crawlRatePerMinute} <= ${input.rateLowThreshold}`,
    };
  }
  return {
    shouldSpawn: false,
    reason: `no pressure (queue=${input.crawlQueueSize}, rate=${input.crawlRatePerMinute})`,
  };
}
