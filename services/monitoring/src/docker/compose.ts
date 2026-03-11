import path from "node:path";
import { config, MONITORING_ROOT } from "../config/env.js";

export type ComposeRef = keyof typeof composeMap;

// Paths in config are relative to the monitoring service root (where .env lives)
export function resolveComposePath(rel: string): string {
  return path.isAbsolute(rel) ? rel : path.resolve(MONITORING_ROOT, rel);
}

export const composeMap = {
  infra: resolveComposePath(config.ROOT_COMPOSE_FILE),
  engine: resolveComposePath(config.ENGINE_COMPOSE_FILE),
  spider: resolveComposePath(config.SPIDER_COMPOSE_FILE),
  indexer: resolveComposePath(config.INDEXER_COMPOSE_FILE),
  ranking: resolveComposePath(config.RANKING_COMPOSE_FILE),
} as const;

export type ServiceRef = {
  compose: string;
  service: string;
};

export const services = {
  psql: { compose: composeMap.infra, service: "psql" },
  redis: { compose: composeMap.infra, service: "redis" },
  adminer: { compose: composeMap.infra, service: "adminer" },
  engine: { compose: composeMap.engine, service: "engine" },
  spider: { compose: composeMap.spider, service: "spider" },
  indexer: { compose: composeMap.indexer, service: "indexer" },
  ranking: { compose: composeMap.ranking, service: "ranking" },
} satisfies Record<string, ServiceRef>;

export type KnownService = keyof typeof services;
