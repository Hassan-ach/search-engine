# Spider Service

Web crawler that discovers and downloads pages.

**Language**: Go
**Purpose**: Crawl websites, extract links, store HTML

## Quick Start

```bash
go mod tidy
go run cmd/spider/main.go
```

Press CTRL+C to stop gracefully.

## Files

- `cmd/spider/main.go` - Entry point with signal handling
- `internal/spider/spider.go` - Core crawling logic
- `internal/parser/parser.go` - HTML parsing and link extraction
- `internal/store/db.go` - Database operations

## Configuration

See root `.env.example` or set environment variables:

```env
SPIDER_MAX_CRAWLERS=5
SPIDER_MAX_CONCURRENT_FETCH=10
SPIDER_HTTP_TIMEOUT=30
SPIDER_CRAWL_DELAY=100
SPIDER_LOGS_PATH=./logs
```

## Output

Stores to PostgreSQL:
- `urls` table - Discovered URLs
- `pages` table - HTML content
- `graph_edges` table - Link relationships

## Details

See [docs/SPIDER.md](../../docs/SPIDER.md) for how it works, database schema, and troubleshooting.
