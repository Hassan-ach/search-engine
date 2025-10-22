CREATE EXTENSION IF NOT EXISTS pgcrypto;

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

CREATE UNIQUE INDEX idx_words_word ON words(word);

CREATE UNIQUE INDEX idx_pages_url ON pages(url);

CREATE INDEX idx_page_word_page_id ON page_word(page_id);

CREATE INDEX idx_page_word_word_id ON page_word(word_id);

CREATE
OR REPLACE FUNCTION recalc_idf() RETURNS TRIGGER AS
$$
DECLARE
total_pages INTEGER;

pages_with_word INTEGER;

new_idf DOUBLE PRECISION;

BEGIN
SELECT
    COUNT(*) INTO total_pages
FROM
    pages;

SELECT
    COUNT(DISTINCT page_id) INTO pages_with_word
FROM
    page_word
WHERE
    word_id = NEW.word_id;

IF pages_with_word > 0 THEN new_idf := LOG(
    total_pages::DOUBLE PRECISION / pages_with_word::DOUBLE PRECISION
) + 1;

UPDATE
    words
SET
    idf = new_idf
WHERE
    id = NEW.word_id;

END IF;

RETURN NEW;

END;

$$
LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS recalc_idf_trigger ON page_word;

CREATE TRIGGER recalc_idf_trigger
AFTER
INSERT
    OR DELETE ON page_word FOR EACH ROW EXECUTE PROCEDURE recalc_idf();
