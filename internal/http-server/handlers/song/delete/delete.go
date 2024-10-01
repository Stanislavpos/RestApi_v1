package delete

import (
	resp "RestApi_v1/internal/config/internal/lib/api/response"
	"RestApi_v1/internal/config/internal/lib/logger/sl"
	"RestApi_v1/internal/config/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"log/slog"
	"net/http"
)

type SongDelete interface {
	DeleteSong(songName string) (string, error)
}

func New(log *slog.Logger, songDelete SongDelete) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		song := chi.URLParam(r, "song")

		if song == "" {
			log.Info("song is empty")

			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		resSong, err := songDelete.DeleteSong(song)
		if errors.Is(err, storage.ErrSongNotFound) {
			log.Info("song not found", "song", song)
			render.JSON(w, r, resp.Error("not found"))
			return
		}
		if err != nil {
			log.Error(op, "failed to delete song", sl.Err(err))
			render.JSON(w, r, resp.Error("internal server error"))
			return
		}

		//var req Request

		log.Info("delete song", slog.String("song", resSong))
		render.JSON(w, r, resSong)

	}
}
