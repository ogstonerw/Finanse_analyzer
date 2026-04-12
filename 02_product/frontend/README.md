# Frontend

React frontend MVP для `Market Reaction Analytics Platform`. Клиентская часть остается тонким слоем поверх существующего backend API и показывает ключевые пользовательские сценарии платформы: вход, обзор рынка, активы, детали инструмента, ленту новостей и событий, а также последние прогнозы.

## Структура экранов

- `LoginPage` - вход по email и password в защищенный аналитический workspace.
- `DashboardPage` - главный desktop-first экран с кризисометром, общей сводкой, KPI, активами в фокусе, событиями и последними прогнозами.
- `AssetsPage` - каталог активов первой версии с поиском, фильтрами и едиными карточками.
- `AssetDetailsPage` - карточка инструмента с ценовым рядом, индикаторами, текущим сигналом, режимом рынка и связанными событиями.
- `NewsEventsPage` - объединенная лента новостей и структурированных событий с локальной фильтрацией.
- `ForecastsPage` - список актуальных прогнозов с выделением выбранного сигнала и кратким объяснением.

## Визуальные принципы

- Единая темная desktop-first палитра собрана по `00_Foundations/color_system.svg`: основной фон, поверхности, акцентный синий, положительные, негативные и информационные состояния.
- Типографика опирается на `00_Foundations/typography.svg`: крупные заголовки для hero и page header, секционные заголовки для карточек, компактные подписи для метрик и таблиц.
- Отступы, радиусы и тени следуют `00_Foundations/spacing_radius_shadows.svg`: scale `4-48`, радиусы `12 / 16 / 20`, карточный shadow для всех ключевых блоков.
- Карточная структура и reusable-блоки собраны в духе `01_Components/components_overview.svg`: sidebar, topbar search, filter chips, KPI cards, content cards, badges, loading и empty states.
- Если backend не возвращает данных, frontend показывает понятные `loading`, `empty` и `error` состояния без добавления новой бизнес-логики.

## Как frontend соотносится с SVG Figma-пакетом

- `02_Desktop/D_Login.svg` -> `LoginPage` и `LoginForm`.
- `02_Desktop/D_Dashboard.svg` -> `DashboardPage`, `RegimeBlock`, `DashboardSummaryBlock`, `ForecastsBlock`, `EventsListBlock`.
- `02_Desktop/D_Assets.svg` -> `AssetsPage`, `AssetCard`, toolbar с поиском и фильтрами.
- `02_Desktop/D_AssetDetails.svg` -> `AssetDetailsPage`, `PricesBlock`, `IndicatorsBlock`, правый сайдбар со сводкой и историей.
- `02_Desktop/D_NewsEvents.svg` -> `NewsEventsPage` и структура event feed.
- `02_Desktop/D_Forecasts.svg` -> `ForecastsPage` и список сигналов в table-like раскладке.
- `03_States/states_overview.svg` -> единые hover, loading, empty и selected states внутри всех экранов.

Реализация не пытается придумать новый UI поверх MVP, а адаптирует текущую React-структуру под предоставленные desktop references и foundations.

## Структура каталога

- `src/pages` - маршрутные экраны MVP.
- `src/components` - повторно используемые карточки, layout-блоки и состояния.
- `src/api` - минимальный клиент для работы с backend.
- `src/lib` - локальное форматирование, storage и общие helpers.

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

Для создания первого пользователя по-прежнему используется backend endpoint:

- `POST /api/v1/auth/register`

## Запуск

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

## Ограничения MVP

- frontend сохраняет пользовательскую сессию локально и не реализует production-ready auth flow;
- registration UI отсутствует;
- сложные графические библиотеки не используются, поэтому ценовой график реализован легким SVG-спарклайном;
- часть экранов использует агрегированные данные из `dashboard summary`, если отдельные endpoint'ы истории пока не предусмотрены;
- приоритет сделан на наглядную демонстрацию архитектуры из `02_product/docs/architecture.md`, а не на полный production UX.
