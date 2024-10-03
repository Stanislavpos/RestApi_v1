package get

import (
	resp "RestApi_v1/internal/config/internal/lib/api/response"
	"RestApi_v1/internal/config/internal/lib/logger/sl"
	"RestApi_v1/internal/config/internal/models"
	"RestApi_v1/internal/config/internal/storage"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strconv"
)

// интерфейс для получения песни с пагинацией
type SongGetter interface {
	GetSongWithPagination(ctx context.Context, id int, page int, pageSize int) ([]models.Song, error)
}

func New(log *slog.Logger, songGetter SongGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.get.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil || id <= 0 {
			log.Info("invalid song id", "id", id)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		page, err := strconv.Atoi(chi.URLParam(r, "page"))
		if err != nil || page <= 0 {
			log.Info("invalid page number", "page", page)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		pageSize, err := strconv.Atoi(chi.URLParam(r, "pageSize"))
		if err != nil || pageSize <= 0 {
			log.Info("invalid page size", "pageSize", pageSize)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		songs, err := songGetter.GetSongWithPagination(r.Context(), id, page, pageSize)
		if errors.Is(err, storage.ErrSongNotFound) {
			log.Info("song not found", "song id", id)
			render.JSON(w, r, resp.Error("not found"))
			return
		}
		if err != nil {
			log.Error(op, "failed to get song", sl.Err(err))
			render.JSON(w, r, resp.Error("internal server error"))
			return
		}

		log.Info("got songs", "songs", songs)
		render.JSON(w, r, songs)
	}
}
