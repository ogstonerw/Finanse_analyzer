# Backend

Минимальный backend для платформы анализа и прогнозирования реакции фондового рынка с первой рабочей вертикалью авторизации.

## Что включено

- точка входа `cmd/api/main.go`;
- каркас HTTP API на стандартной библиотеке Go;
- подключение к PostgreSQL через `database/sql` и драйвер `github.com/lib/pq`;
- загрузка конфигурации из переменных окружения и примера `.env`;
- миграции для таблиц `users` и `user_sessions`;
- storage/repository слой для пользователей и сессий;
- модули `auth`, `users`, `assets`, `forecasts`, `regime`, `storage`;
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

В каталоге `migrations/` добавлены начальные SQL-миграции:

- `001_create_users.sql`
- `002_create_user_sessions.sql`

На текущем этапе они применяются любым удобным инструментом миграций или вручную через `psql`.

## Запуск

```bash
cd 02_product/backend
go mod tidy
go run ./cmd/api
```

## Базовые маршруты

- `POST /api/v1/auth/register` — регистрация пользователя по `email` и `password`
- `POST /api/v1/auth/login` — вход пользователя по `email` и `password`
- `GET /api/v1/assets` — список активов-заглушек
- `GET /api/v1/forecasts/latest` — последняя прогнозная запись-заглушка
- `GET /api/v1/regime/current` — текущее состояние режима рынка-заглушка

## Статус

Реализована первая рабочая вертикаль `auth + users/user_sessions`: пароль хранится только в виде hash, при регистрации и входе создается пользовательская сессия. Refresh-flow, email verification, сложная валидация и расширенная бизнес-логика пока не реализованы.
