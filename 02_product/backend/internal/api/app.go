package api

import (
	"net/http"
	"time"

	"diploma-market-ai/02_product/backend/internal/assets"
	"diploma-market-ai/02_product/backend/internal/auth"
	"diploma-market-ai/02_product/backend/internal/events"
	"diploma-market-ai/02_product/backend/internal/forecasts"
	"diploma-market-ai/02_product/backend/internal/indicators"
	"diploma-market-ai/02_product/backend/internal/news"
	"diploma-market-ai/02_product/backend/internal/prices"
	"diploma-market-ai/02_product/backend/internal/regime"
	"diploma-market-ai/02_product/backend/internal/storage"
	"diploma-market-ai/02_product/backend/internal/users"
)

type App struct {
	config            Config
	router            *http.ServeMux
	authHandler       *auth.Handler
	assetsHandler     *assets.Handler
	indicatorsHandler *indicators.Handler
	pricesHandler     *prices.Handler
	sourcesHandler    *SourcesHandler
	newsHandler       *news.Handler
	eventsHandler     *events.Handler
	forecastsHandler  *forecasts.Handler
	regimeHandler     *regime.Handler
}

func NewApp(
	cfg Config,
	store *storage.Postgres,
	pricesService *prices.Service,
	indicatorsService *indicators.Service,
	newsService *news.Service,
	eventsService *events.Service,
	forecastsService *forecasts.Service,
) *App {
	userRepo := users.NewRepository(store)
	authService := auth.NewService(userRepo, time.Duration(cfg.SessionTTLHours)*time.Hour)
	assetsService := assets.NewService(store)
	sourcesRepository := storage.NewSourcesRepository(store)
	regimeService := regime.NewService(store)

	app := &App{
		config:            cfg,
		router:            http.NewServeMux(),
		authHandler:       auth.NewHandler(authService),
		assetsHandler:     assets.NewHandler(assetsService),
		indicatorsHandler: indicators.NewHandler(indicatorsService),
		pricesHandler:     prices.NewHandler(pricesService),
		sourcesHandler:    NewSourcesHandler(sourcesRepository),
		newsHandler:       news.NewHandler(newsService),
		eventsHandler:     events.NewHandler(eventsService),
		forecastsHandler:  forecasts.NewHandler(forecastsService),
		regimeHandler:     regime.NewHandler(regimeService),
	}

	app.registerRoutes()

	return app
}

func (a *App) Router() http.Handler {
	return a.router
}

func (a *App) registerRoutes() {
	a.router.HandleFunc("/healthz", a.handleHealth)
	a.router.HandleFunc("POST /api/v1/auth/register", a.authHandler.Register)
	a.router.HandleFunc("POST /api/v1/auth/login", a.authHandler.Login)
	a.router.HandleFunc("GET /api/v1/assets", a.assetsHandler.List)
	a.router.HandleFunc("GET /api/v1/assets/{ticker}", a.assetsHandler.GetByTicker)
	a.router.HandleFunc("GET /api/v1/assets/{ticker}/indicators", a.indicatorsHandler.ListByTicker)
	a.router.HandleFunc("GET /api/v1/assets/{ticker}/prices", a.pricesHandler.ListByTicker)
	a.router.HandleFunc("GET /api/v1/sources", a.sourcesHandler.List)
	a.router.HandleFunc("GET /api/v1/news", a.newsHandler.List)
	a.router.HandleFunc("GET /api/v1/news/{id}", a.newsHandler.GetByID)
	a.router.HandleFunc("GET /api/v1/events", a.eventsHandler.List)
	a.router.HandleFunc("POST /api/v1/forecasts/generate", a.forecastsHandler.Generate)
	a.router.HandleFunc("GET /api/v1/forecasts/latest", a.forecastsHandler.Latest)
	a.router.HandleFunc("GET /api/v1/regime/current", a.regimeHandler.Current)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"env":    a.config.Environment,
	})
}
