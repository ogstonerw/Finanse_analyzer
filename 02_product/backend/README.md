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

## Как backend используется во frontend MVP

Frontend экранов `login`, `dashboard`, `assets`, `asset details`, `news/events` и `forecasts` опирается на backend так:

- `login` -> `POST /api/v1/auth/login`
- `dashboard` -> `GET /api/v1/dashboard/summary`, `GET /api/v1/regime/current`
- `assets` -> `GET /api/v1/assets` с дополнительным enrichment из `GET /api/v1/dashboard/summary`
- `asset details` -> `GET /api/v1/assets/{ticker}`, `GET /api/v1/assets/{ticker}/prices`, `GET /api/v1/assets/{ticker}/indicators`, enrichment из `GET /api/v1/dashboard/summary`
- `news/events` -> `GET /api/v1/news`, `GET /api/v1/events`
- `forecasts` -> `GET /api/v1/forecasts/latest`, `GET /api/v1/dashboard/summary`

Для демонстрации это важно: frontend теперь выдерживает частичную недоступность вторичных endpoint'ов и показывает понятные `loading`, `empty`, `error` и warning-состояния, но базовый сценарий всё равно зависит от запущенного backend и подготовленной БД.

## Подготовка к запуску

Перед первым запуском нужно вручную подготовить PostgreSQL:

1. Создать базу данных `market_ai`, если она еще не создана.
2. Применить миграции `02_product/backend/migrations/001-010` через `psql` или pgAdmin.
3. Выставить переменные окружения:
   - `DB_HOST`
   - `DB_PORT`
   - `DB_USER`
   - `DB_PASSWORD`
   - `DB_NAME`
   - `DB_SSLMODE`

Встроенного migration runner в backend пока нет, поэтому без ручного применения миграций сервис запустится, но начнет отдавать ошибки уровня БД по отсутствующим таблицам.

## Запуск backend

```bash
cd 02_product/backend
go run ./cmd/api
```

При старте backend выполняет начальную синхронизацию цен, индикаторов, новостей и событий. На пустой или неподготовленной БД это приведет к warning-логам, поэтому миграции должны быть применены заранее.

## Минимальный сценарий демонстрации

1. Подготовить PostgreSQL и применить миграции.
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
