# Backend Skeleton

Минимальный backend skeleton для платформы анализа и прогнозирования реакции фондового рынка.

## Что включено

- точка входа `cmd/api/main.go`;
- каркас HTTP API на стандартной библиотеке Go;
- загрузка конфигурации из переменных окружения;
- заготовка PostgreSQL-конфигурации и слоя хранения;
- модули `auth`, `assets`, `forecasts`, `regime`, `storage`;
- минимальные маршруты:
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `GET /api/v1/assets`
  - `GET /api/v1/forecasts/latest`
  - `GET /api/v1/regime/current`

## Переменные окружения

- `APP_HOST` — хост API, по умолчанию `0.0.0.0`
- `APP_PORT` — порт API, по умолчанию `8080`
- `APP_ENV` — окружение, по умолчанию `development`
- `DB_HOST` — хост PostgreSQL, по умолчанию `localhost`
- `DB_PORT` — порт PostgreSQL, по умолчанию `5432`
- `DB_USER` — пользователь PostgreSQL, по умолчанию `postgres`
- `DB_PASSWORD` — пароль PostgreSQL, по умолчанию `postgres`
- `DB_NAME` — имя базы данных, по умолчанию `market_ai`
- `DB_SSLMODE` — режим SSL, по умолчанию `disable`

## Запуск

```bash
cd 02_product/backend
go run ./cmd/api
```

## Статус

Skeleton не содержит полной бизнес-логики, миграций, реального подключения драйвера PostgreSQL и прикладных сценариев работы с БД. На текущем этапе он задает базовую структуру backend-приложения для дальнейшей реализации.
