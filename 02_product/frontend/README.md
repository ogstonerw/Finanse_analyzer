# Frontend

React frontend MVP для `Market Reaction Analytics Platform`. Клиентская часть остается тонким слоем поверх существующего backend API и ориентирована на стабильную и понятную демонстрацию перед предзащитой.

## Экранная структура

- `LoginPage` - вход по email и password.
- `DashboardPage` - обзор рынка, кризисометр, summary, KPI, активы, события и прогнозы.
- `AssetsPage` - список активов первой версии с поиском и фильтрами.
- `AssetDetailsPage` - карточка инструмента с ценами, индикаторами, прогнозом и связанными событиями.
- `NewsEventsPage` - объединенная лента новостей и событий.
- `ForecastsPage` - список актуальных прогнозов и детальная карточка выбранного сигнала.

## Что стабилизировано для демо

- для всех экранов предусмотрены `loading states`;
- для пустых выборок предусмотрены `empty states`;
- для сетевых и backend-ошибок предусмотрены `error states` с повторным запросом;
- для частичной недоступности вторичных endpoint'ов предусмотрены warning-блоки без падения всего экрана;
- заголовки, подписи и статусные labels приведены к более понятному демо-формату.

## Как frontend работает с backend endpoint'ами

- `login` -> `POST /api/v1/auth/login`
- `dashboard` -> `GET /api/v1/dashboard/summary`, `GET /api/v1/regime/current`
- `assets` -> `GET /api/v1/assets`, enrichment из `GET /api/v1/dashboard/summary`
- `asset details` -> `GET /api/v1/assets/{ticker}`, `GET /api/v1/assets/{ticker}/prices`, `GET /api/v1/assets/{ticker}/indicators`, enrichment из `GET /api/v1/dashboard/summary`
- `news/events` -> `GET /api/v1/news`, `GET /api/v1/events`
- `forecasts` -> `GET /api/v1/forecasts/latest`, `GET /api/v1/dashboard/summary`

Если secondary endpoint временно не отвечает, frontend старается сохранить основной экран работоспособным и явно показывает, какой именно блок деградировал.

## Как frontend соотносится с локальным design package

- `02_Desktop/D_Login.svg` -> `LoginPage`, `LoginForm`
- `02_Desktop/D_Dashboard.svg` -> `DashboardPage`, `RegimeBlock`, `DashboardSummaryBlock`
- `02_Desktop/D_Assets.svg` -> `AssetsPage`, `AssetCard`
- `02_Desktop/D_AssetDetails.svg` -> `AssetDetailsPage`, `PricesBlock`, `IndicatorsBlock`
- `02_Desktop/D_NewsEvents.svg` -> `NewsEventsPage`
- `02_Desktop/D_Forecasts.svg` -> `ForecastsPage`
- `03_States/states_overview.svg` -> loading, empty, error и selected states

## Запуск

1. Сначала запустить backend из `02_product/backend`.
2. Затем запустить frontend:

```bash
cd 02_product/frontend
npm install
npm run dev
```

По умолчанию Vite поднимает dev server на `http://localhost:5173`.

## Подключение к backend

Для локальной разработки в `vite.config.js` включен proxy на backend `http://localhost:8080`, поэтому запросы к `/api/*` работают без отдельной CORS-настройки.

При необходимости можно явно задать backend URL:

```bash
VITE_API_BASE_URL=http://localhost:8080
```

Если backend недоступен, frontend теперь показывает понятные error states с подсказкой проверить локальный запуск сервера.

## Минимальный сценарий демонстрации

1. Убедиться, что PostgreSQL подготовлена и backend запущен.
2. Если пользователя нет, создать его через `POST /api/v1/auth/register`.
3. Открыть `http://localhost:5173/login`.
4. Войти в систему.
5. Последовательно показать:
   - `dashboard`
   - `assets`
   - `asset details`
   - `news/events`
   - `forecasts`

## Ограничения MVP

- frontend сохраняет пользовательскую сессию локально и не реализует production-ready auth flow;
- registration UI отсутствует;
- ценовой график реализован простым SVG-спарклайном без chart library;
- часть экранов использует агрегированные данные из `dashboard summary`;
- приоритет сделан на устойчивость демонстрации и понятность связки frontend с backend, а не на полный production UX.
