export interface IndexerPolicyInput {
  unindexedPages: number;
  runningIndexerJobs: number;
  lastSpawnedAt: number | null; // epoch ms
  nowMs: number;
  threshold: number;
  maxParallel: number;
  cooldownMs: number;
}

export interface IndexerPolicyDecision {
  shouldSpawn: boolean;
  reason: string;
}

export function decideIndexer(
  input: IndexerPolicyInput
): IndexerPolicyDecision {
  if (input.unindexedPages < input.threshold) {
    return {
      shouldSpawn: false,
      reason: `backlog ${input.unindexedPages} below threshold ${input.threshold}`,
    };
  }
  if (input.runningIndexerJobs >= input.maxParallel) {
    return {
      shouldSpawn: false,
      reason: `already at max parallel ${input.maxParallel}`,
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
  return {
    shouldSpawn: true,
    reason: `backlog ${input.unindexedPages} >= threshold, spawning indexer`,
  };
}
