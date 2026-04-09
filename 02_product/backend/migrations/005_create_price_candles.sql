CREATE TABLE IF NOT EXISTS price_candles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    timeframe VARCHAR(20) NOT NULL,
    candle_time TIMESTAMP NOT NULL,
    open_price NUMERIC(18,6) NOT NULL,
    high_price NUMERIC(18,6) NOT NULL,
    low_price NUMERIC(18,6) NOT NULL,
    close_price NUMERIC(18,6) NOT NULL,
    volume NUMERIC(20,6) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (asset_id, timeframe, candle_time)
);

CREATE INDEX IF NOT EXISTS idx_price_candles_asset_timeframe_time
    ON price_candles (asset_id, timeframe, candle_time DESC);
