CREATE TABLE IF NOT EXISTS news_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    external_id VARCHAR(255) NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    body TEXT,
    published_at TIMESTAMP NOT NULL,
    collected_at TIMESTAMP NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (source_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_news_items_published_at
    ON news_items (published_at DESC);

CREATE INDEX IF NOT EXISTS idx_news_items_source_published_at
    ON news_items (source_id, published_at DESC);
