# Backend

Минимальный backend платформы анализа и прогнозирования реакции фондового рынка. Текущий MVP включает базовую авторизацию, справочники `assets` и `sources`, загрузку дневных свечей из MOEX ISS, расчёт технических индикаторов, а также контур `news_items` и `events` на официальном источнике Банка России.

## Что включено

- точка входа `cmd/api/main.go`;
- HTTP API на стандартной библиотеке Go;
- PostgreSQL через `database/sql` и драйвер `github.com/lib/pq`;
- конфигурация через переменные окружения;
- SQL-миграции для `users`, `user_sessions`, `assets`, `sources`, `price_candles`, `technical_indicators`, `news_items`, `events`;
- storage/repository слой для пользователей, сессий, активов, источников, свечей, индикаторов, новостей и событий;
- модули `auth`, `users`, `assets`, `prices`, `indicators`, `news`, `events`, `collectors`, `forecasts`, `regime`, `storage`.

## Сущности MVP

- `assets` хранит базовый перечень целевых инструментов и внешних факторов MVP.
- `sources` хранит зафиксированные источники данных проекта.
- `price_candles` хранит дневные исторические свечи по поддерживаемым активам.
- `technical_indicators` хранит рассчитанные признаки по дневным свечам.
- `news_items` хранит нормализованные новости: `title`, `summary/body`, `source_id`, `published_at`, `collected_at`, `url`.
- `events` хранит базовые события, извлечённые из новостей: `asset_id` при наличии связи, `event_type`, `summary`, `extracted_at`.

## Источники данных

- Исторические цены загружаются из официального MOEX ISS.
- Для MVP-новостей используется официальный источник `Bank of Russia Monetary Policy Decisions` с `https://www.cbr.ru/eng/dkp/mp_dec/`.
- При старте backend выполняет базовую синхронизацию свечей, затем расчёт индикаторов, затем загрузку новостей и rule-based извлечение событий.

## Индикаторы

По дневным свечам рассчитываются:

- `RSI(14)`
- недельная доходность
- историческая волатильность
- направление тренда
- положение цены в локальном диапазоне

Если истории недостаточно, числовые поля остаются `null`, а состояние фиксируется через `calculation_status`.

## News и Events

- Новости нормализуются в единую структуру без сложного NLP.
- Для MVP используется простая классификация событий по ключевым словам.
- Возможные типы включают `key_rate_cut`, `key_rate_hold`, `key_rate_hike`, `monetary_policy`, `general_news` и несколько базовых категорий для корпоративных и сырьевых новостей.
- Если новость не удаётся явно связать с активом, `asset_id` в событии остаётся `null`.

## Переменные окружения

- `APP_HOST` — хост API, по умолчанию `0.0.0.0`
- `APP_PORT` — порт API, по умолчанию `8080`
- `APP_ENV` — окружение, по умолчанию `development`
- `APP_SESSION_TTL_HOURS` — срок жизни пользовательской сессии в часах, по умолчанию `24`
- `DB_HOST` — хост PostgreSQL, по умолчанию `localhost`
- `DB_PORT` — порт PostgreSQL, по умолчанию `5432`
- `DB_USER` — пользователь PostgreSQL, по умолчанию `postgres`
- `DB_PASSWORD` — пароль PostgreSQL, по умолчанию `postgres`
- `DB_NAME` — имя базы данных, по умолчанию `market_ai`
- `DB_SSLMODE` — режим SSL, по умолчанию `disable`
- `DB_MAX_OPEN_CONNS` — максимум открытых соединений, по умолчанию `10`
- `DB_MAX_IDLE_CONNS` — максимум idle-соединений, по умолчанию `5`
- `DB_CONN_MAX_LIFETIME_MINUTES` — время жизни соединения, по умолчанию `30`

Пример значений вынесен в `configs/app.env.example`.

## Миграции

В каталоге `migrations/` находятся:

- `001_create_users.sql`
- `002_create_user_sessions.sql`
- `003_create_assets.sql`
- `004_create_sources.sql`
- `005_create_price_candles.sql`
- `006_create_technical_indicators.sql`
- `007_create_news_items.sql`
- `008_create_events.sql`

## Запуск

```bash
cd 02_product/backend
go mod tidy
go run ./cmd/api
```

## Основные маршруты

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/assets`
- `GET /api/v1/assets/{ticker}`
- `GET /api/v1/assets/{ticker}/prices`
- `GET /api/v1/assets/{ticker}/indicators`
- `GET /api/v1/sources`
- `GET /api/v1/news`
- `GET /api/v1/news/{id}`
- `GET /api/v1/events`
- `GET /api/v1/forecasts/latest`
- `GET /api/v1/regime/current`

## Статус

Реализована рабочая основа backend: авторизация, справочники, исторические цены, технические индикаторы, MVP-контур новостей и базовых событий. Планировщик, realtime-обновление, sentiment scoring, importance scoring, AI-интеграция и сложная прогнозная логика пока не реализованы.
