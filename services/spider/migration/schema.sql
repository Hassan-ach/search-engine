CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE pages (
    url TEXT PRIMARY KEY,
    url_hash BYTEA GENERATED ALWAYS AS (digest(url, 'sha256')) STORED,
    html TEXT NOT NULL,
    metadata JSONB NOT NULL
);

CREATE UNIQUE INDEX idx_pages_url_hash ON pages (url_hash);

