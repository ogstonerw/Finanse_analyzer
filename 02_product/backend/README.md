# Backend

Минимальный backend платформы анализа и прогнозирования реакции фондового рынка. Текущий MVP включает авторизацию, справочники `assets` и `sources`, загрузку дневных свечей из MOEX ISS, расчет технических индикаторов, контур `news_items` и `events`, а также первый pipeline генерации комбинированного прогноза в таблицу `forecasts`.

## Что включено

- точка входа `cmd/api/main.go`;
- HTTP API на стандартной библиотеке Go;
- PostgreSQL через `database/sql` и драйвер `github.com/lib/pq`;
- конфигурация через переменные окружения;
- SQL-миграции для `users`, `user_sessions`, `assets`, `sources`, `price_candles`, `technical_indicators`, `news_items`, `events`, `forecasts`;
- storage/repository слой для пользователей, сессий, активов, источников, свечей, индикаторов, новостей, событий и прогнозов;
- модули `auth`, `users`, `assets`, `prices`, `indicators`, `news`, `events`, `ai`, `collectors`, `forecasts`, `regime`, `storage`.

## Сущности MVP

- `assets` хранит базовый перечень целевых инструментов и внешних факторов.
- `sources` хранит каталог источников данных проекта.
- `price_candles` хранит дневные исторические свечи.
- `technical_indicators` хранит рассчитанные признаки по дневным свечам.
- `news_items` хранит нормализованные новости.
- `events` хранит базовые события, извлеченные из новостей.
- `forecasts` хранит рассчитанные комбинированные сигналы прогноза с пояснением, режимом AI и рыночным контекстом.

## Источники данных

- Исторические цены загружаются из официального MOEX ISS.
- Для MVP-новостей используется официальный источник `Bank of Russia Monetary Policy Decisions`.
- При старте backend выполняет базовую синхронизацию свечей, затем расчет индикаторов, затем загрузку новостей и извлечение событий.

## Индикаторы

По дневным свечам рассчитываются:

- `RSI(14)`
- недельная доходность
- историческая волатильность
- направление тренда
- положение цены в локальном диапазоне

Если истории недостаточно, числовые поля остаются `null`, а полнота расчета отмечается через `calculation_status`.

## News и Events

- Новости приводятся к единой структуре: `title`, `summary`, `body`, `source_id`, `published_at`, `collected_at`, `url`.
- Для MVP используется rule-based классификация событий по ключевым словам и тикерам.
- Если новость не удается явно связать с активом, `asset_id` в событии остается `null`.

## Forecast Pipeline

- `POST /api/v1/forecasts/generate` принимает `ticker` и необязательный `event_id`.
- Если `event_id` не передан, backend использует последний релевантный event для актива или общий event без привязки к активу.
- Сервис объединяет:
  - event и связанную news;
  - asset;
  - последние `technical_indicators` по активу;
  - общий рыночный контекст на основе последних индикаторов `IMOEX`.
- Результат сохраняется в `forecasts` и возвращает:
  - `direction`
  - `strength`
  - `confidence`
  - `explanation`

## AI Режимы

Модуль `internal/ai` работает в двух режимах:

- `fallback` — прогноз формируется rule-based логикой без внешнего API.
- `prepare` — backend все еще возвращает fallback-результат, но дополнительно собирает структурированный payload для будущего внешнего AI API.

На текущем этапе внешний вызов OpenAI API не выполняется. Режим `prepare` нужен как прозрачная точка расширения под следующую итерацию.

## Переменные окружения

- `APP_HOST` — хост API, по умолчанию `0.0.0.0`
- `APP_PORT` — порт API, по умолчанию `8080`
- `APP_ENV` — окружение, по умолчанию `development`
- `APP_SESSION_TTL_HOURS` — TTL пользовательской сессии в часах, по умолчанию `24`
- `AI_MODE` — режим AI-модуля: `fallback` или `prepare`
- `AI_PROVIDER` — логический провайдер внешнего AI, по умолчанию `openai`
- `AI_MODEL` — имя будущей внешней модели
- `AI_API_ENDPOINT` — endpoint будущего внешнего AI API
- `AI_API_KEY` — ключ внешнего AI API
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
- `009_create_forecasts.sql`

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
- `POST /api/v1/forecasts/generate`
- `GET /api/v1/forecasts/latest`
- `GET /api/v1/regime/current`

## Статус

Реализована рабочая основа backend: авторизация, справочники, исторические цены, технические индикаторы, контур новостей и событий, а также первый MVP-контур комбинированного прогноза с fallback AI-режимом. Реальная внешняя AI-интеграция, scheduler, batch processing, sentiment scoring и расширенная кризисометрия пока не реализованы.
