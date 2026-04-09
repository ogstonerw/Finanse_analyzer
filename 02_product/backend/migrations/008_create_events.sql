CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    news_item_id UUID NOT NULL UNIQUE REFERENCES news_items(id) ON DELETE CASCADE,
    asset_id UUID REFERENCES assets(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    summary TEXT NOT NULL,
    extracted_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_extracted_at
    ON events (extracted_at DESC);

CREATE INDEX IF NOT EXISTS idx_events_asset_extracted_at
    ON events (asset_id, extracted_at DESC);
