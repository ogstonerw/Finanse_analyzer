CREATE TABLE IF NOT EXISTS technical_indicators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    indicator_time TIMESTAMP NOT NULL,
    timeframe VARCHAR(20) NOT NULL,
    weekly_return NUMERIC(10,6),
    rsi NUMERIC(10,4),
    volatility NUMERIC(10,6),
    trend_direction VARCHAR(20),
    channel_position NUMERIC(10,4),
    calculation_status VARCHAR(50) NOT NULL DEFAULT 'insufficient_data',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (asset_id, timeframe, indicator_time)
);

CREATE INDEX IF NOT EXISTS idx_technical_indicators_asset_timeframe_time
    ON technical_indicators (asset_id, timeframe, indicator_time DESC);
