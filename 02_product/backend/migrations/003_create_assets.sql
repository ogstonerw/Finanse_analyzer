CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    asset_type VARCHAR(50) NOT NULL,
    sector VARCHAR(100) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO assets (ticker, name, asset_type, sector, currency, is_active)
VALUES
    ('IMOEX', 'Индекс Московской биржи', 'index', 'market_index', 'RUB', TRUE),
    ('SBER', 'Сбер', 'equity', 'banking', 'RUB', TRUE),
    ('LKOH', 'Лукойл', 'equity', 'oil_gas', 'RUB', TRUE),
    ('GAZP', 'Газпром', 'equity', 'oil_gas', 'RUB', TRUE),
    ('YDEX', 'Яндекс', 'equity', 'technology', 'RUB', TRUE),
    ('BRENT', 'Нефть Brent', 'commodity', 'energy', 'USD', TRUE),
    ('NATGAS', 'Природный газ', 'commodity', 'energy', 'USD', TRUE)
ON CONFLICT (ticker) DO NOTHING;
