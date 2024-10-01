package save

import (
	resp "RestApi_v1/internal/config/internal/lib/api/response"
	"RestApi_v1/internal/config/internal/lib/logger/sl"
	"RestApi_v1/internal/config/internal/storage"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Request struct {
	Song     string     `json:"song" validate:"required"`
	Group    string     `json:"group,omitempty"`
	TextSong string     `json:"text_song,omitempty"`
	DateSong *time.Time `json:"date_song,omitempty"`
	LinkSong string     `json:"link_song,omitempty"`
}

type Response struct {
	resp.Response
}

type SongSaver interface {
	SaveSong(SongToSave string, GroupToSave string, TextSongToSave string, DateToSave time.Time, LinkToSave string) (int64, error)
}

func New(log *slog.Logger, songSaver SongSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty")
				render.JSON(w, r, resp.Error("empty request"))
			} else {
				log.Error("failed to decode request body", sl.Err(err))
				render.JSON(w, r, resp.Error("failed to decode request"))
			}
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			if validateErr, ok := err.(validator.ValidationErrors); ok {
				log.Error("invalid request", sl.Err(err))
				render.JSON(w, r, resp.ValidationError(validateErr))
			} else {
				log.Error("unexpected error during validation", sl.Err(err))
				render.JSON(w, r, resp.Error("validation error"))
			}
			return
		}

		// Устанавливаем стандартное значение для даты, если не указана
		dateToSave := time.Now() // или time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC) по вашей логике
		if req.DateSong != nil {
			dateToSave = *req.DateSong
		}

		log.Info("Saving song", slog.String("song", req.Song), slog.String("group", req.Group), slog.String("textSong", req.TextSong), slog.String("linkSong", req.LinkSong))
		id, err := songSaver.SaveSong(req.Song, req.Group, req.TextSong, dateToSave, req.LinkSong)
		if errors.Is(err, storage.ErrSongExist) {
			log.Info("song already exists", slog.String("song", req.Song))
			render.JSON(w, r, resp.Error("song already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add song", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to add song"))
			return
		}

		log.Info("song added", slog.Int64("id", id))
		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
