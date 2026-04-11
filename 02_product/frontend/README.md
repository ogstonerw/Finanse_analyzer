# Frontend

MVP frontend платформы на React для демонстрации backend API на предзащите. Интерфейс показывает login-форму, dashboard, список активов, карточку актива, новости и события, а также последние прогнозы и блок кризисометра.

## Что внутри

- `src/pages` — страницы приложения:
  - `LoginPage`
  - `DashboardPage`
  - `AssetsPage`
  - `AssetDetailsPage`
  - `NewsEventsPage`
  - `ForecastsPage`
- `src/components` — повторно используемые MVP-блоки интерфейса:
  - `LoginForm`
  - `DashboardSummaryBlock`
  - `AssetCard`
  - `PricesBlock`
  - `IndicatorsBlock`
  - `ForecastsBlock`
  - `RegimeBlock`
- `src/api` — минимальный клиент для запросов к backend.
- `src/lib` — форматирование дат, чисел и вспомогательные функции.

## Backend endpoints

Frontend использует уже существующие backend endpoint'ы:

- `POST /api/v1/auth/login`
- `GET /api/v1/dashboard/summary`
- `GET /api/v1/regime/current`
- `GET /api/v1/assets`
- `GET /api/v1/assets/{ticker}`
- `GET /api/v1/assets/{ticker}/prices`
- `GET /api/v1/assets/{ticker}/indicators`
- `GET /api/v1/news`
- `GET /api/v1/events`
- `GET /api/v1/forecasts/latest`

## Запуск

```bash
cd 02_product/frontend
npm install
npm run dev
```

По умолчанию Vite поднимает dev server на `http://localhost:5173`.

## Подключение к backend

Для локальной разработки в `vite.config.js` включен proxy на backend `http://localhost:8080`, поэтому запросы к `/api/*` работают без отдельной CORS-настройки.

Если нужен явный URL backend, можно задать:

```bash
VITE_API_BASE_URL=http://localhost:8080
```

## Ограничения MVP

- нет полноценной защищенной сессии на фронте;
- нет chart libraries;
- нет сложного глобального state management;
- нет production-ready UX, optimistic updates и расширенной обработки ошибок.

Текущая версия предназначена для понятной демонстрации структуры платформы и уже готовых backend возможностей.
