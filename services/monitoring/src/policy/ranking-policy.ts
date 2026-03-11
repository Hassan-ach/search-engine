export interface RankingPolicyInput {
  totalIndexedPages: number;
  lastRankingAt: bigint | null; // stored as total indexed at time of last run
  triggerDelta: number;
  runningRankingJobs: number;
  maxParallel: number;
}

export interface RankingPolicyDecision {
  shouldRun: boolean;
  reason: string;
}

export function decideRanking(
  input: RankingPolicyInput
): RankingPolicyDecision {
  if (input.runningRankingJobs >= input.maxParallel) {
    return {
      shouldRun: false,
      reason: `ranking already running (${input.runningRankingJobs})`,
    };
  }
  const baseline = Number(input.lastRankingAt ?? 0n);
  const delta = input.totalIndexedPages - baseline;
  if (delta < input.triggerDelta) {
    return {
      shouldRun: false,
      reason: `delta ${delta} < trigger ${input.triggerDelta}`,
    };
  }
  return {
    shouldRun: true,
    reason: `delta ${delta} >= trigger ${input.triggerDelta}, running ranking`,
  };
}
