# Services Overview

Five services implementing a complete search engine pipeline.

## Pipeline

```
Web → Spider → Indexer → Ranking → Engine → User
                              ↑
                         Monitoring (watches & restarts all)
```

## Services

| Service | Language | Purpose |
|---------|----------|---------|
| [Engine](engine/) | Go | Web interface, search, results |
| [Spider](spider/) | Go | Web crawler, discovery, links |
| [Indexer](indexer/) | Rust | HTML parsing, word extraction, TF |
| [Ranking](ranking/) | Python | Scoring (TF-IDF, PageRank) |
| [Monitoring](monitoring/) | TypeScript | Service orchestration, health checks, live dashboard |

## Quick Start

From root:
```bash
docker-compose up
open http://localhost:1323
```

## Each Service

See individual READMEs for:
- [Engine](engine/README.md) - Web server
- [Spider](spider/README.md) - Web crawler
- [Indexer](indexer/README.md) - Text indexing
- [Ranking](ranking/README.md) - Scoring algorithms
- [Monitoring](monitoring/README.md) - Orchestration and dashboard

## Detailed Docs

Advanced information in [docs/](../docs/):
- [docs/ENGINE.md](../docs/ENGINE.md) - Architecture, templates, configuration
- [docs/SPIDER.md](../docs/SPIDER.md) - Crawling algorithm, database schema
- [docs/INDEXER.md](../docs/INDEXER.md) - TF calculation, word extraction
- [docs/RANKING.md](../docs/RANKING.md) - PageRank math, iterative scoring
- [docs/MONITORING.md](../docs/MONITORING.md) - Tick lifecycle, policy engine, metrics API

## Data Flow

```
Spider → pages, urls, graph_edges tables
          ↓
        Indexer → words, page_word tables
          ↓
        Ranking → Updates words (IDF), Creates page_rank
          ↓
        Engine ← Queries words, page_rank for results
```

## Setup

1. Copy root `.env.example` to `.env`
2. Customize if needed
3. Run `docker-compose up`

See [root README](../README.md) for details.
