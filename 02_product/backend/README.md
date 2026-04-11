# Backend

Минимальный backend платформы анализа и прогнозирования реакции фондового рынка. Текущий MVP включает авторизацию, справочники `assets` и `sources`, загрузку дневных свечей, расчет технических индикаторов, контур `news_items` и `events`, генерацию прогнозов в `forecasts`, а также rule-based контур `market_regime`, кризисометра и `dashboard summary`.

## Что реализовано

- HTTP API на стандартной библиотеке Go.
- PostgreSQL через `database/sql` и `github.com/lib/pq`.
- SQL-миграции `001`-`010` для `users`, `user_sessions`, `assets`, `sources`, `price_candles`, `technical_indicators`, `news_items`, `events`, `forecasts`, `market_regime`.
- Storage/repository слой для активов, источников, свечей, индикаторов, новостей, событий, прогнозов и режима рынка.
- Модули `auth`, `assets`, `prices`, `indicators`, `news`, `events`, `forecasts`, `regime`, `storage`.

## Market Regime и Crisisometer

`GET /api/v1/regime/current` рассчитывает и сохраняет общий режим рынка на основе уже доступных данных:

- рыночный стресс по последним индикаторам `IMOEX`;
- новостный стресс по последним событиям;
- макроэкономический стресс по событиям денежно-кредитного фона;
- сырьевой стресс по доступным commodity-индикаторам или commodity-событиям;
- breadth stress по состоянию ключевых активов.

Ответ включает:

- `regime_score`;
- `regime_label`;
- `sub_scores`;
- `summary`;
- `explanation`;
- `calculation_model`.

Текущая реализация является временным MVP-контуром с `calculation_model = rule_based_mvp`. Scheduler, realtime, исторический расчет `market_regime` и сложная аналитика пока не реализованы.

## Dashboard Summary

`GET /api/v1/dashboard/summary` возвращает агрегированную сводку для frontend:

- текущий `regime`;
- список активов с последними индикаторами;
- последние прогнозы;
- последние события;
- краткий общий `summary`.

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

## Подготовка к запуску

Перед первым запуском нужно вручную подготовить PostgreSQL:

1. создать базу данных `market_ai`, если она еще не создана;
2. применить миграции `02_product/backend/migrations/001-010` через `psql` или pgAdmin;
3. выставить переменные окружения `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`.

Встроенного migration runner в backend пока нет, поэтому без ручного применения миграций сервис будет запускаться, но начнет отдавать ошибки уровня БД по отсутствующим таблицам.

## Запуск

```bash
cd 02_product/backend
go run ./cmd/api
```

При старте backend выполняет начальную синхронизацию цен, индикаторов, новостей и событий. На пустой или неподготовленной БД это приведет к warning-логам, поэтому миграции должны быть применены заранее.

## Авторизация и frontend

- backend поддерживает и регистрацию, и логин через `POST /api/v1/auth/register` и `POST /api/v1/auth/login`;
- текущий React frontend содержит только login-экран;
- первого пользователя для frontend нужно создать через backend endpoint регистрации.

## Локальная проверка

```bash
go build -buildvcs=false ./...
go test -buildvcs=false ./...
```
