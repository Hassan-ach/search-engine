import { query } from "./client.js";

// ── Ensure monitor_state table exists ────────────────────────────────────────
export async function ensureMonitorStateTable(): Promise<void> {
  await query(`
    CREATE TABLE IF NOT EXISTS monitor_state (
      key TEXT PRIMARY KEY,
      value BIGINT NOT NULL,
      updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    )
  `);
}

// ── Indexer metrics ───────────────────────────────────────────────────────────
export async function countUnindexedPages(): Promise<number> {
  const result = await query<{ count: string }>(
    "SELECT COUNT(*) AS count FROM pages WHERE indexed = false"
  );
  return parseInt(result.rows[0]?.count ?? "0", 10);
}

export async function countIndexedPagesTotal(): Promise<number> {
  const result = await query<{ count: string }>(
    "SELECT COUNT(*) AS count FROM pages WHERE indexed = true"
  );
  return parseInt(result.rows[0]?.count ?? "0", 10);
}

// ── Spider metrics ────────────────────────────────────────────────────────────
export async function countPagesCrawledLastMinute(): Promise<number> {
  const result = await query<{ count: string }>(
    "SELECT COUNT(*) AS count FROM pages WHERE created_at >= NOW() - INTERVAL '1 minute'"
  );
  return parseInt(result.rows[0]?.count ?? "0", 10);
}

// ── Cleanup ───────────────────────────────────────────────────────────────────
export async function cleanupIndexedHtml(batchSize: number): Promise<number> {
  const result = await query<{ count: string }>(
    `WITH to_update AS (
       SELECT id FROM pages
       WHERE indexed = true AND html <> ''
       LIMIT $1
     )
     UPDATE pages
     SET html = '', updated_at = NOW()
     WHERE id IN (SELECT id FROM to_update)
     RETURNING id`,
    [batchSize]
  );
  return result.rowCount ?? 0;
}

// ── Monitor state key-value store ─────────────────────────────────────────────
export async function monitorStateGet(key: string): Promise<bigint | null> {
  const result = await query<{ value: string }>(
    "SELECT value FROM monitor_state WHERE key = $1",
    [key]
  );
  if (!result.rows[0]) return null;
  return BigInt(result.rows[0].value);
}

export async function monitorStateSet(
  key: string,
  value: bigint
): Promise<void> {
  await query(
    `INSERT INTO monitor_state (key, value, updated_at)
     VALUES ($1, $2, NOW())
     ON CONFLICT (key) DO UPDATE
       SET value = EXCLUDED.value, updated_at = NOW()`,
    [key, value.toString()]
  );
}
