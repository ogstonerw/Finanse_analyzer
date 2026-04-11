# Session 12

## Цель

Собрать React MVP frontend для платформы, подключить его к уже существующим backend endpoint'ам и подготовить удобный демонстрационный интерфейс для предзащиты.

## Что сделано

- создан новый каталог `02_product/frontend/` с минимальным React + Vite приложением;
- добавлена базовая маршрутизация для страниц:
  - `login`
  - `dashboard`
  - `assets list`
  - `asset details`
  - `news/events list`
  - `latest forecasts`
- реализован простой локальный MVP-контур авторизации через `POST /api/v1/auth/login`;
- подключены backend endpoint'ы:
  - `GET /api/v1/dashboard/summary`
  - `GET /api/v1/regime/current`
  - `GET /api/v1/assets`
  - `GET /api/v1/assets/{ticker}`
  - `GET /api/v1/assets/{ticker}/prices`
  - `GET /api/v1/assets/{ticker}/indicators`
  - `GET /api/v1/news`
  - `GET /api/v1/events`
  - `GET /api/v1/forecasts/latest`
- реализованы MVP-компоненты:
  - `LoginForm`
  - `DashboardSummaryBlock`
  - `AssetCard`
  - `PricesBlock`
  - `IndicatorsBlock`
  - `ForecastsBlock`
  - `RegimeBlock`
- добавлен общий `styles.css` с простым светлым интерфейсом, пригодным для демонстрации без сложного дизайна;
- создан `package-lock.json` для воспроизводимой установки зависимостей;
- обновлен `02_product/frontend/README.md` с описанием структуры, запуска и используемых endpoint'ов.

## Измененные файлы

- `02_product/frontend/.gitignore`
- `02_product/frontend/index.html`
- `02_product/frontend/package.json`
- `02_product/frontend/package-lock.json`
- `02_product/frontend/vite.config.js`
- `02_product/frontend/README.md`
- `02_product/frontend/src/main.jsx`
- `02_product/frontend/src/App.jsx`
- `02_product/frontend/src/styles.css`
- `02_product/frontend/src/api/client.js`
- `02_product/frontend/src/lib/authStorage.js`
- `02_product/frontend/src/lib/formatters.js`
- `02_product/frontend/src/lib/useRemoteData.js`
- `02_product/frontend/src/components/AppLayout.jsx`
- `02_product/frontend/src/components/AssetCard.jsx`
- `02_product/frontend/src/components/DashboardSummaryBlock.jsx`
- `02_product/frontend/src/components/EmptyState.jsx`
- `02_product/frontend/src/components/EventsListBlock.jsx`
- `02_product/frontend/src/components/ForecastsBlock.jsx`
- `02_product/frontend/src/components/IndicatorsBlock.jsx`
- `02_product/frontend/src/components/LoadingBlock.jsx`
- `02_product/frontend/src/components/LoginForm.jsx`
- `02_product/frontend/src/components/PageHeader.jsx`
- `02_product/frontend/src/components/PricesBlock.jsx`
- `02_product/frontend/src/components/RegimeBlock.jsx`
- `02_product/frontend/src/pages/LoginPage.jsx`
- `02_product/frontend/src/pages/DashboardPage.jsx`
- `02_product/frontend/src/pages/AssetsPage.jsx`
- `02_product/frontend/src/pages/AssetDetailsPage.jsx`
- `02_product/frontend/src/pages/NewsEventsPage.jsx`
- `02_product/frontend/src/pages/ForecastsPage.jsx`
- `05_session_logs/session_12.md`

## Следующий логичный шаг

Добавить небольшой слой фильтров и пользовательских сценариев для демо: выбор тикера из интерфейса, обновление данных кнопками и более явное связывание forecasts с текущим market regime на фронте.
