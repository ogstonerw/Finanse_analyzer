CREATE TABLE IF NOT EXISTS market_regime (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID REFERENCES assets(id) ON DELETE SET NULL,
    regime_time TIMESTAMP NOT NULL,
    regime_label VARCHAR(50) NOT NULL,
    regime_score NUMERIC(5,2) NOT NULL,
    market_stress_score NUMERIC(5,2) NOT NULL,
    news_stress_score NUMERIC(5,2) NOT NULL,
    macro_stress_score NUMERIC(5,2) NOT NULL,
    commodity_stress_score NUMERIC(5,2) NOT NULL,
    breadth_stress_score NUMERIC(5,2) NOT NULL,
    summary TEXT NOT NULL,
    explanation TEXT NOT NULL,
    calculation_model VARCHAR(50) NOT NULL DEFAULT 'rule_based_mvp',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_market_regime_regime_time
    ON market_regime (regime_time DESC);

CREATE INDEX IF NOT EXISTS idx_market_regime_asset_time
    ON market_regime (asset_id, regime_time DESC);
