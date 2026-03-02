# Indexer Service

Converts crawled HTML into searchable word index.

**Language**: Rust  
**Purpose**: Extract words, calculate TF, build search index

## Quick Start

```bash
cargo build
cargo run
```

Press CTRL+C to stop gracefully.

## Files

- `src/main.rs` - Entry point, task spawning
- `src/core/indexer.rs` - Indexing algorithm
- `src/core/psql.rs` - Database operations
- `src/core/text_sink.rs` - Word extraction

## Configuration

```env
INDEXER_COUNT=4         # Worker count
INDEXER_LOG_PATH=./logs # Log file location
```

## Input/Output

**Reads from**:
- `pages` table (where indexed = false)

**Writes to**:
- `words` table
- `page_word` table (with TF values)

## Details

See [docs/INDEXER.md](../../docs/INDEXER.md) for algorithm explanation, performance details, and troubleshooting.

## Notes

⚠️ Learning project - simple tokenization, English-only.
