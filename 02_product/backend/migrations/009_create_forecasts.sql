CREATE TABLE IF NOT EXISTS forecasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    event_id UUID REFERENCES events(id) ON DELETE SET NULL,
    forecast_horizon VARCHAR(50) NOT NULL,
    forecast_time TIMESTAMP NOT NULL,
    direction_label VARCHAR(50) NOT NULL,
    signal_strength NUMERIC(5,2) NOT NULL,
    confidence_score NUMERIC(5,2) NOT NULL,
    explanation TEXT NOT NULL,
    ai_mode VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    market_context_label VARCHAR(50) NOT NULL,
    market_context_score NUMERIC(5,2) NOT NULL,
    market_context_explanation TEXT NOT NULL,
    key_factors_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    prepared_request_json JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_forecasts_forecast_time
    ON forecasts (forecast_time DESC);

CREATE INDEX IF NOT EXISTS idx_forecasts_asset_forecast_time
    ON forecasts (asset_id, forecast_time DESC);

CREATE INDEX IF NOT EXISTS idx_forecasts_event_id
    ON forecasts (event_id);
