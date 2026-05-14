# Текущий статус проекта

## Готово

- зафиксирована тема проекта, состав целевых активов и недельный горизонт прогноза;
- согласована трехглавная структура ВКР и подготовлены базовые проектные документы;
- описаны архитектура платформы, логика прогнозирования, логика кризисометра и логическая схема БД;
- реализован backend MVP на Go с PostgreSQL и SQL-миграциями `001`-`010`;
- в backend доступны контуры:
  - `auth`,
  - `assets`,
  - `sources`,
  - `price_candles`,
  - `technical_indicators`,
  - `news_items`,
  - `events`,
  - `forecasts`,
  - `market_regime`;
- реализованы API endpoint'ы для авторизации, активов, свечей, индикаторов, новостей, событий, прогнозов, режима рынка и dashboard summary;
- как этап `session_11` реализованы:
  - MVP-контур кризисометра,
  - таблица и repository-слой `market_regime`,
  - `GET /api/v1/regime/current`,
  - `GET /api/v1/dashboard/summary`,
  - временная прозрачная модель `rule_based_mvp`;
- как этап `session_12` реализован React frontend MVP:
  - `login`,
  - `dashboard`,
  - `assets list`,
  - `asset details`,
  - `news/events list`,
  - `latest forecasts`;
- frontend подключен к backend API через Vite proxy и использует уже реализованные backend endpoint'ы;
- backend README, frontend README и журналы `session_11` и `session_12` синхронизированы с фактическим состоянием репозитория.
- как этап `session_14` начата интеграция AI-агента по первому варианту:
  - добавлен режим `AI_MODE=openai` для обращения к OpenAI Responses API;
  - сохранен режим `fallback` для локальной демонстрации без внешнего API;
  - добавлена строгая JSON-схема результата прогноза;
  - подготовленный AI payload сохраняется без API-ключа.
- как этап `session_15` усилен демонстрационный контур MVP:
  - backend применяет SQL-миграции при старте через `ApplyMigrations`;
  - frontend согласован с backend-направлениями `up`, `down`, `neutral`;
  - на странице прогнозов добавлен запуск `POST /api/v1/forecasts/generate`;
  - установлены Go/gofmt и Node/npm для локальной проверки;
  - backend tests и frontend build проходят локально.

## Фактическое состояние MVP

- backend и frontend находятся в репозитории и готовы к локальному запуску;
- PostgreSQL-схема в репозитории реализуется через миграции `02_product/backend/migrations/001-010`;
- backend применяет SQL-миграции при старте из каталога `DB_MIGRATIONS_DIR`, по умолчанию `migrations`;
- при старте backend выполняет начальную синхронизацию цен, индикаторов, новостей и событий;
- frontend пока содержит только login-экран, без отдельной страницы регистрации;
- создание пользователя выполняется через backend endpoint `POST /api/v1/auth/register`;
- scheduler, realtime, исторический пересчет `market_regime`, расширенная аналитика, chart libraries и production-ready UX пока не реализованы.

## Важное уточнение по документации

- `02_product/docs/database_schema.md` остается логическим проектным описанием целевой схемы БД уровня ВКР;
- фактическая MVP-реализация базы данных определяется SQL-миграциями в `02_product/backend/migrations/`;
- при описании реализованного состояния проекта для backend и frontend следует ориентироваться именно на миграции, README и журналы сессий `session_11` и `session_12`.

## Следующий логичный шаг

- проверить полный demo-flow на локальной БД:
  - регистрация первого пользователя через backend endpoint;
  - вход во frontend;
  - генерация прогноза в режиме `fallback`;
  - при наличии API-ключа отдельная проверка `AI_MODE=openai`;
  - затем подготовить описание реализации для разделов 3.1-3.3 ВКР.
