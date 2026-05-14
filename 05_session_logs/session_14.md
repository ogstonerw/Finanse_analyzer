# Session 14

## Этап

Этап `session_14` фиксирует начало интеграции AI-агента по первому варианту: встроенный OpenAI-клиент внутри Go backend без выделения отдельного сервиса.

## Цель

Подключить режим `AI_MODE=openai` к существующему контуру `forecasts`, сохранив `fallback` для локальной демонстрации и воспроизводимости MVP.

## Что сделано

- уточнена документация логики недельного сигнала;
- добавлен режим `openai` в `internal/ai.Client`;
- добавлена отправка запроса в OpenAI Responses API;
- добавлена строгая JSON-схема результата прогноза:
  - `direction`,
  - `strength`,
  - `confidence`,
  - `explanation`,
  - `key_factors`;
- добавлена нормализация результата перед сохранением прогноза;
- подготовленный AI payload сохраняется без API-ключа;
- расширен backend config переменными:
  - `AI_REASONING_EFFORT`,
  - `AI_TEXT_VERBOSITY`,
  - `AI_TIMEOUT_SECONDS`;
- добавлены unit-тесты для OpenAI-клиента с mock HTTP server.

## Измененные файлы

- `00_rules/current_status.md`
- `00_rules/next_task.md`
- `02_product/backend/README.md`
- `02_product/backend/cmd/api/main.go`
- `02_product/backend/configs/app.env.example`
- `02_product/backend/internal/ai/client.go`
- `02_product/backend/internal/ai/client_test.go`
- `02_product/backend/internal/api/config.go`
- `02_product/docs/forecast_logic.md`
- `05_session_logs/session_14.md`

## Проверка

- `where.exe go` - Go не найден в текущем `PATH`.
- `gofmt -w ...` - не выполнено, так как `gofmt` не найден в текущем `PATH`.
- `go test -buildvcs=false ./...` - не выполнено, так как `go` не найден в текущем `PATH`.

## Следующий логичный шаг

Исправить frontend-согласование направлений `up/down/neutral` и добавить запуск генерации прогноза из интерфейса.
