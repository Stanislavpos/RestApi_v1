package main

import (
	"RestApi_v1/internal/config/internal/config"
	updateSong "RestApi_v1/internal/config/internal/http-server/handlers/song/updateSong"

	del "RestApi_v1/internal/config/internal/http-server/handlers/song/delete"
	"RestApi_v1/internal/config/internal/http-server/handlers/song/get"
	"RestApi_v1/internal/config/internal/http-server/handlers/song/save"
	"RestApi_v1/internal/config/internal/lib/logger/sl"
	"RestApi_v1/internal/config/internal/storage/postgres"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting service", slog.String("env", cfg.Env))
	log.Debug("debug msg are enable")

	storage, err := postgres.New()
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)

	router.Use(middleware.Logger)

	router.Use(middleware.Recoverer)

	router.Post("/song", save.New(log, storage))
	router.Get("/{id}/{page}/{pageSize}", get.New(log, storage))

	router.Delete("/{song}", del.New(log, storage))
	router.Put("/edit", updateSong.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.Timeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
