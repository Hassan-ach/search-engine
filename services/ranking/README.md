# Ranking Service

Calculates page importance & relevance scores.

**Language**: Python  
**Algorithms**: TF-IDF, PageRank

## Quick Start

```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python src/main.py
```

Press CTRL+C to stop.

## Files

- `src/main.py` - Entry point, scheduling
- `src/idf.py` - Inverse document frequency calculation
- `src/page_rank.py` - PageRank algorithm
- `src/psql.py` - Database operations

## Configuration

```env
RANKING_DAMPING_FACTOR=0.85      # 85% follow links, 15% random
RANKING_ITERATIONS=20            # Convergence iterations
RANKING_SCHEDULE_INTERVAL=3600   # Update every hour
```

## Input/Output

**Reads from**:
- `words`, `page_word` tables (for TF-IDF)
- `graph_edges` table (for PageRank)

**Writes to**:
- `words` table (IDF values)
- `page_rank` table (page scores)

## Details

See [docs/RANKING.md](../../docs/RANKING.md) for algorithm details, math formulas, performance tuning, and examples.

## Notes

⚠️ Learning project - no optimization, simple iteration-based PageRank.
