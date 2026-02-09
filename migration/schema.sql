CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE urls (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    -- status TEXT NOT NULL CHECK (status IN ('pending', 'crawled', 'failed')),
);

CREATE TABLE pages (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    html TEXT NOT NULL,
    metadata JSONB NOT NULL,
    indexed BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE words (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    word VARCHAR(25) UNIQUE NOT NULL,
    idf DOUBLE PRECISION NOT NULL DEFAULT 1
);

CREATE TABLE page_word (
    page_id UUID NOT NULL,
    word_id UUID NOT NULL,
    tf INTEGER NOT NULL,
    PRIMARY KEY (page_id, word_id),
    FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE,
    FOREIGN KEY (word_id) REFERENCES words(id) ON DELETE CASCADE
);

CREATE TABLE graph_edges (
    id BIGSERIAL PRIMARY KEY,
    from_page UUID NOT NULL,
    to_page   UUID NOT NULL,
    UNIQUE (from_page, to_page)
    FOREIGN KEY (from_page) REFERENCES urls(id) ON DELETE CASCADE,
    FOREIGN KEY (to_page) REFERENCES urls(id) ON DELETE CASCADE
);

CREATE TABLE page_rank (
    page_id UUID PRIMARY KEY REFERENCES pages(id),
    score   DOUBLE PRECISION NOT NULL
);

-- CREATE TABLE image_page (
--     image_url TEXT NOT NULL,
--     page_id  UUID NOT NULL,
--     PRIMARY KEY (image_url, page_id),
--     FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
-- );



CREATE UNIQUE INDEX idx_words_word ON words(word);

CREATE UNIQUE INDEX idx_pages_url ON pages(url);

CREATE INDEX idx_page_word_page_id ON page_word(page_id);

CREATE INDEX idx_page_word_word_id ON page_word(word_id);

CREATE INDEX idx_graph_edges_from_page ON graph_edges(from_page);

CREATE INDEX idx_graph_edges_to_page ON graph_edges(to_page);

CREATE INDEX idx_page_rank_score ON page_rank(score DESC);

CREATE INDEX idx_urls_url ON urls(url);

-- CREATE INDEX idx_image_page_image_url ON image_page(image_url);
--
-- CREATE INDEX idx_image_page_page_id ON image_page(page_id);
