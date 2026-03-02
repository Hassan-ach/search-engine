# Boogle - Learning Search Engine

A 2000s-style search engine built for educational purposes.

## Quick Start

```bash
# Start all services
docker-compose up

# Open browser
open http://localhost:1323
```

## Services

- **Engine** ([services/engine/](services/engine/)) - Web interface & query handler
- **Spider** ([services/spider/](services/spider/)) - Web crawler
- **Indexer** ([services/indexer/](services/indexer/)) - Converts HTML to searchable index
- **Ranking** ([services/ranking/](services/ranking/)) - Calculates page scores (TF-IDF, PageRank)

## Requirements

- Docker & Docker Compose
- PostgreSQL (via docker-compose)
- Redis (via docker-compose)

## Setup

1. Copy `.env.example` to `.env` and customize if needed
2. Run `docker-compose up`
3. Visit http://localhost:1323

## Docs

- [Services Overview](services/README.md) - Service details
- [docs/](docs/) - Detailed documentation for each service

## Development

See individual service READMEs in [services/](services/) directory.

---

**Status**: Learning project - not for production
**License**: MIT
