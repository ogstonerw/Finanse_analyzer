CREATE TABLE IF NOT EXISTS sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    base_url TEXT NOT NULL UNIQUE,
    source_type VARCHAR(50) NOT NULL,
    access_method VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    update_frequency VARCHAR(100) NOT NULL,
    last_checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO sources (
    name,
    base_url,
    source_type,
    access_method,
    status,
    update_frequency,
    last_checked_at
)
VALUES
    ('Moscow Exchange ISS', 'https://www.moex.com/a8531', 'market', 'http_api', 'primary', 'В течение торгового дня, исторические данные доступны по запросу', TIMESTAMP '2026-04-07 00:00:00'),
    ('Moscow Exchange Interfaces', 'https://www.moex.com/a7939', 'market', 'web', 'reference', 'По мере обновления биржевой инфраструктуры', TIMESTAMP '2026-04-07 00:00:00'),
    ('Moscow Exchange Indices', 'https://www.moex.com/en/indices/', 'index', 'web', 'primary', 'В течение торгового дня и по графику пересмотров составов', TIMESTAMP '2026-04-07 00:00:00'),
    ('MOEX Russia Index (IMOEX)', 'https://www.moex.com/a6231', 'index', 'web', 'primary_context', 'В течение торгового дня', TIMESTAMP '2026-04-07 00:00:00'),
    ('Bank of Russia Monetary Policy Decisions', 'https://www.cbr.ru/eng/dkp/mp_dec/', 'macro', 'web', 'primary', 'По датам заседаний совета директоров', TIMESTAMP '2026-04-07 00:00:00'),
    ('Bank of Russia Key Rate', 'https://www.cbr.ru/eng/hd_base/KeyRate/', 'macro', 'web', 'primary', 'После изменения ставки, исторический ряд поддерживается постоянно', TIMESTAMP '2026-04-07 00:00:00'),
    ('Calendar of Key Rate Decisions', 'https://www.cbr.ru/eng/dkp/cal_mp/', 'macro', 'web', 'reference', 'По годовому календарю Банка России', TIMESTAMP '2026-04-07 00:00:00'),
    ('Rosstat Advance Release Calendar', 'https://eng.rosstat.gov.ru/folder/13943', 'macro', 'web', 'primary', 'По календарю Росстата', TIMESTAMP '2026-04-07 00:00:00'),
    ('Rosstat Prices and Inflation', 'https://www.rosstat.gov.ru/statistics/price/', 'macro', 'web', 'primary', 'Еженедельно и ежемесячно в зависимости от показателя', TIMESTAMP '2026-04-07 00:00:00'),
    ('Rosstat National Accounts', 'https://eng.rosstat.gov.ru/folder/13913', 'macro', 'web', 'primary', 'Квартально и ежегодно', TIMESTAMP '2026-04-07 00:00:00'),
    ('Bank of Russia Statistics', 'https://www.cbr.ru/eng/statistics/', 'macro', 'web', 'primary', 'По публикационному календарю ЦБ', TIMESTAMP '2026-04-07 00:00:00'),
    ('Bank of Russia Official Statistics Release Calendar', 'https://www.cbr.ru/eng/statistics/indcalendar', 'macro', 'web', 'reference', 'Ежедневно, еженедельно, ежемесячно по типу показателя', TIMESTAMP '2026-04-07 00:00:00'),
    ('Центр раскрытия корпоративной информации Интерфакс', 'https://www.e-disclosure.ru/', 'corporate', 'web', 'primary', 'По мере публикации эмитентами', TIMESTAMP '2026-04-07 00:00:00'),
    ('E-disclosure Поиск по сообщениям', 'https://e-disclosure.ru/poisk-po-soobshheniyam', 'corporate', 'web', 'primary', 'По мере публикации эмитентами', TIMESTAMP '2026-04-07 00:00:00'),
    ('E-disclosure Шлюз API', 'https://e-disclosure.ru/poluchenie-informacii/shlyuz-api', 'corporate', 'api_gateway', 'planned', 'По мере публикации эмитентами', TIMESTAMP '2026-04-07 00:00:00'),
    ('E-disclosure FTP выгрузка', 'https://e-disclosure.ru/poluchenie-informacii/vygruzka-na-ftp', 'corporate', 'ftp', 'planned', 'По мере публикации и обновления выгрузок', TIMESTAMP '2026-04-07 00:00:00'),
    ('EIA Spot Prices for Crude Oil and Petroleum Products', 'https://www.eia.gov/dnav/pet/PET_PRI_SPT_S1_D.htm', 'commodity', 'web', 'primary', 'Ежедневно', TIMESTAMP '2026-04-07 00:00:00'),
    ('EIA Henry Hub Natural Gas Spot Price', 'https://www.eia.gov/dnav/ng/hist/rngwhhdA.htm', 'commodity', 'web', 'primary', 'Ежедневно, еженедельно, ежемесячно и ежегодно в зависимости от режима просмотра', TIMESTAMP '2026-04-07 00:00:00'),
    ('EIA Natural Gas Weekly Update', 'https://www.eia.gov/naturalgas/weekly/', 'commodity', 'web', 'secondary', 'Еженедельно', TIMESTAMP '2026-04-07 00:00:00'),
    ('Интерфакс Экономика', 'https://www.interfax.ru/business/', 'news', 'web', 'secondary', 'В течение дня', TIMESTAMP '2026-04-07 00:00:00'),
    ('ТАСС Экономика и бизнес', 'https://tass.ru/ekonomika', 'news', 'web', 'secondary', 'В течение дня', TIMESTAMP '2026-04-07 00:00:00')
ON CONFLICT (base_url) DO NOTHING;
