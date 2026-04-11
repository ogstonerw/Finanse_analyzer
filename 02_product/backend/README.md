# Backend

Минимальный backend платформы анализа и прогнозирования реакции фондового рынка. Текущий MVP включает авторизацию, справочники `assets` и `sources`, загрузку дневных свечей, расчет технических индикаторов, контур `news_items` и `events`, генерацию прогнозов в `forecasts`, а также rule-based контур `market_regime` и кризисометра.

## Что реализовано

- HTTP API на стандартной библиотеке Go.
- PostgreSQL через `database/sql` и `github.com/lib/pq`.
- SQL-миграции для `users`, `user_sessions`, `assets`, `sources`, `price_candles`, `technical_indicators`, `news_items`, `events`, `forecasts`, `market_regime`.
- Storage/repository слой для активов, источников, свечей, индикаторов, новостей, событий, прогнозов и режима рынка.
- Модули `auth`, `assets`, `prices`, `indicators`, `news`, `events`, `forecasts`, `regime`, `storage`.

## MVP сущности

- `forecasts` хранит последние комбинированные прогнозы по активам.
- `market_regime` хранит сохраненные снимки общего состояния рынка.
- Кризисометр в MVP реализован как прозрачная временная rule-based модель с `calculation_model = rule_based_mvp`.

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

Это временная MVP-реализация без scheduler, realtime и сложной аналитики. Дальнейшее развитие предполагает замену rule-based логики на более богатую модель без изменения публичного API-контракта.

## Dashboard Summary

`GET /api/v1/dashboard/summary` возвращает агрегированную сводку для фронта:

- текущий `regime`;
- список активов с последними индикаторами, если они уже рассчитаны;
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

## Запуск

```bash
cd 02_product/backend
go run ./cmd/api
```

Для локальной проверки сборки:

```bash
go build -buildvcs=false ./...
go test -buildvcs=false ./...
```
