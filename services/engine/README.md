# Engine Service

Web server & search interface.

**Language**: Go  
**Framework**: Echo, Templ, Tailwind CSS

## Quick Start

```bash
go mod tidy
npm install
just dev
```

Open http://localhost:1323

## Files

- `cmd/server/main.go` - Entry point
- `internal/handlers/search.handler.go` - Search logic
- `view/` - HTML templates
- `public/` - Compiled assets

## Commands

```bash
just templ-generate  # Generate templates
just tailwind-build  # Build CSS
just js-build        # Build JavaScript
just dev             # Development mode
just build           # Production build
```

## Details

See [docs/ENGINE.md](../../docs/ENGINE.md) for architecture, configuration, and troubleshooting.

