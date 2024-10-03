package main

import (
	//_ "RestApi_v1/internal/config/cmd/restApi_v1/docs"
	"RestApi_v1/internal/config/internal/config"
	del "RestApi_v1/internal/config/internal/http-server/handlers/song/delete"
	"RestApi_v1/internal/config/internal/http-server/handlers/song/get"
	"RestApi_v1/internal/config/internal/http-server/handlers/song/save"
	updateSong "RestApi_v1/internal/config/internal/http-server/handlers/song/updateSong"
	"RestApi_v1/internal/config/internal/lib/logger/sl"
	"RestApi_v1/internal/config/internal/storage/postgres"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/swaggo/swag/example/basic/docs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// gin-swagger middleware
// swagger embed files

// @host localhost:8082
const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	docs.SwaggerInfo.Title = "songs from db API"
	docs.SwaggerInfo.Description = "This is a api for get song from database."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "songs.swagger.io"
	docs.SwaggerInfo.BasePath = "/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	router := chi.NewRouter()

	// Загрузка файла документации
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting service", slog.String("env", cfg.Env))
	log.Debug("debug msg are enable")
	// старт базы данных, создаем таблицу если ее нет
	storage, err := postgres.New() // start db
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router.Use(middleware.RequestID)

	router.Use(middleware.Logger)

	router.Use(middleware.Recoverer)

	// Определение маршрутов
	// Метод Post - добавляем в базу данных песню
	//{
	//	"group": "Название группы2",
	//	"song": "1234",
	//	"textSong": "Текст песни223",
	//	"dateSong": "2023-01-01",
	//	"linkSong": "http://ссылка-на-песня"
	//}
	router.Post("/song", save.New(log, storage)) // add song to db

	// метод Get - получаем из базы данных песню по id, нужно указать номер страницы и размер страницы для пагинации
	router.Get("/{id}/{page}/{pageSize}", get.New(log, storage)) // get song from db
	// метод Delete  - удаляем из базы данных песню имени
	router.Delete("/{song}", del.New(log, storage)) // delete song from db
	// метод Put  - изменяем песню в базе данных, нужно в запросе передать айди - по айди идет поиск в базе
	router.Put("/edit", updateSong.New(log, storage)) // edit song in db

	// Подключение Swagger UI
	router.Get("/swagger/*", httpSwagger.WrapHandler)

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
