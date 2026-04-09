package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"diploma-market-ai/02_product/backend/internal/api"
	"diploma-market-ai/02_product/backend/internal/collectors"
	"diploma-market-ai/02_product/backend/internal/events"
	"diploma-market-ai/02_product/backend/internal/indicators"
	"diploma-market-ai/02_product/backend/internal/news"
	"diploma-market-ai/02_product/backend/internal/prices"
	"diploma-market-ai/02_product/backend/internal/storage"
)

func main() {
	cfg := api.LoadConfig()

	store, err := storage.NewPostgres(storage.Config{
		Host:            cfg.DatabaseHost,
		Port:            cfg.DatabasePort,
		User:            cfg.DatabaseUser,
		Password:        cfg.DatabasePassword,
		Name:            cfg.DatabaseName,
		SSLMode:         cfg.DatabaseSSLMode,
		MaxOpenConns:    cfg.DatabaseMaxOpenConns,
		MaxIdleConns:    cfg.DatabaseMaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.DatabaseConnMaxLifeMin) * time.Minute,
	})
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	pricesService := prices.NewService(store, collectors.NewMOEXCollector(nil))

	priceSyncCtx, priceSyncCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	if err := pricesService.SyncSupportedDailyCandles(priceSyncCtx); err != nil {
		log.Printf("moex price sync warning: %v", err)
	}
	priceSyncCancel()

	indicatorsService := indicators.NewService(store)

	indicatorSyncCtx, indicatorSyncCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	if err := indicatorsService.SyncAllDailyIndicators(indicatorSyncCtx); err != nil {
		log.Printf("technical indicators sync warning: %v", err)
	}
	indicatorSyncCancel()

	newsService := news.NewService(store, collectors.NewCBRMonetaryPolicyCollector(nil))

	newsSyncCtx, newsSyncCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	if err := newsService.SyncLatest(newsSyncCtx); err != nil {
		log.Printf("news sync warning: %v", err)
	}
	newsSyncCancel()

	eventsService := events.NewService(store)

	eventsSyncCtx, eventsSyncCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	if err := eventsService.SyncFromNews(eventsSyncCtx); err != nil {
		log.Printf("events sync warning: %v", err)
	}
	eventsSyncCancel()

	app := api.NewApp(cfg, store, pricesService, indicatorsService, newsService, eventsService)
	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           app.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	defer store.Close()

	go func() {
		log.Printf("backend listening on %s", cfg.Address())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}
