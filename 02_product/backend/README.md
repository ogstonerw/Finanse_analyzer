# Backend

Backend MVP платформы анализа и прогнозирования реакции фондового рынка. Серверная часть реализована на Go и предоставляет API для авторизации, активов, свечей, технических индикаторов, новостей, событий, прогнозов, кризисометра и `dashboard summary`.

## Что реализовано

- HTTP API на стандартной библиотеке Go.
- PostgreSQL через `database/sql` и `github.com/lib/pq`.
- SQL-миграции `001`-`010` для:
  - `users`
  - `user_sessions`
  - `assets`
  - `sources`
  - `price_candles`
  - `technical_indicators`
  - `news_items`
  - `events`
  - `forecasts`
  - `market_regime`
- Storage/repository слой для активов, источников, свечей, индикаторов, новостей, событий, прогнозов и режима рынка.
- Контуры `auth`, `assets`, `prices`, `indicators`, `news`, `events`, `forecasts`, `regime`, `storage`.
- AI-клиент для генерации недельного сигнала в режимах:
  - `fallback` - прозрачная rule-based модель без внешнего API;
  - `prepare` - подготовка payload для проверки контекста;
  - `openai` - обращение к OpenAI Responses API со структурированным JSON-ответом.

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
- `GET /api/v1/dashboard/summary`

## Market Regime и Dashboard Summary

`GET /api/v1/regime/current` возвращает текущий режим рынка и кризисометр на основе уже доступных данных. Ответ включает:

- `regime_score`
- `regime_label`
- `sub_scores`
- `summary`
- `explanation`
- `calculation_model`

Текущая реализация остается rule-based MVP-контуром с `calculation_model = rule_based_mvp`.

`GET /api/v1/dashboard/summary` возвращает агрегированную сводку для frontend:

- текущий `regime`
- список активов с последними индикаторами
- последние прогнозы
- последние события
- краткий общий `summary`

## AI-режимы прогнозирования

`POST /api/v1/forecasts/generate` собирает структурированный контекст прогноза: актив, связанное событие и новость, последние технические индикаторы и текущий рыночный режим. Затем сервис вызывает `internal/ai.Client`.

Для локальной демонстрации по умолчанию используется:

```bash
AI_MODE=fallback
```

Для включения внешнего AI-агента нужно задать:

```bash
AI_MODE=openai
AI_PROVIDER=openai
AI_MODEL=gpt-5.2
AI_API_ENDPOINT=https://api.openai.com/v1/responses
AI_API_KEY=<your_api_key>
AI_TIMEOUT_SECONDS=30
```

В режиме `openai` backend отправляет запрос в Responses API и требует структурированный JSON с полями `direction`, `strength`, `confidence`, `explanation`, `key_factors`. API-ключ не сохраняется в БД; в `prepared_request_json` сохраняется только безопасное описание payload без секрета.

Опционально можно задать `AI_REASONING_EFFORT` и `AI_TEXT_VERBOSITY`, если выбранная модель поддерживает эти параметры. Если они пустые, backend не добавляет их в запрос.

## Как backend используется во frontend MVP

Frontend экранов `login`, `dashboard`, `assets`, `asset details`, `news/events` и `forecasts` опирается на backend так:

- `login` -> `POST /api/v1/auth/login`
- `dashboard` -> `GET /api/v1/dashboard/summary`, `GET /api/v1/regime/current`
- `assets` -> `GET /api/v1/assets` с дополнительным enrichment из `GET /api/v1/dashboard/summary`
- `asset details` -> `GET /api/v1/assets/{ticker}`, `GET /api/v1/assets/{ticker}/prices`, `GET /api/v1/assets/{ticker}/indicators`, enrichment из `GET /api/v1/dashboard/summary`
- `news/events` -> `GET /api/v1/news`, `GET /api/v1/events`
- `forecasts` -> `GET /api/v1/forecasts/latest`, `GET /api/v1/dashboard/summary`, `POST /api/v1/forecasts/generate`

Для демонстрации это важно: frontend теперь выдерживает частичную недоступность вторичных endpoint'ов и показывает понятные `loading`, `empty`, `error` и warning-состояния, но базовый сценарий всё равно зависит от запущенного backend и подготовленной БД.

## Подготовка к запуску

Перед первым запуском нужно подготовить PostgreSQL:

1. Создать базу данных `market_ai`, если она еще не создана.
2. Выставить переменные окружения:
   - `DB_HOST`
   - `DB_PORT`
   - `DB_USER`
   - `DB_PASSWORD`
   - `DB_NAME`
   - `DB_SSLMODE`
   - `DB_MIGRATIONS_DIR`

Backend применяет SQL-миграции из `DB_MIGRATIONS_DIR` при старте. По умолчанию используется каталог `migrations`, поэтому стандартный запуск из `02_product/backend` не требует ручного применения файлов `001-010`.

## Запуск backend

```bash
cd 02_product/backend
go run ./cmd/api
```

При старте backend сначала применяет миграции, затем выполняет начальную синхронизацию цен, индикаторов, новостей и событий. Если внешние источники временно недоступны, сервис продолжит запуск с warning-логами.

## Минимальный сценарий демонстрации

1. Подготовить PostgreSQL.
2. Запустить backend.
3. Если пользователя еще нет, создать его через `POST /api/v1/auth/register`.
4. Запустить frontend из `02_product/frontend`.
5. Открыть `http://localhost:5173/login`.
6. Последовательно показать:
   - вход
   - dashboard
   - assets
   - asset details
   - news/events
   - forecasts
   - формирование нового прогноза через кнопку `Сформировать прогноз`

## Локальная проверка

```bash
go build -buildvcs=false ./...
go test -buildvcs=false ./...
```

Для frontend после backend-запуска используется:

```bash
cd ../frontend
npm install
npm run dev
```
