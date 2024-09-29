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
)

type Request struct {
	Song     string `json:"song" validate:"required,song"`
	Group    string `json:"group"`
	TextSong string `json:"text_song"`
	DateSong string `json:"date_song"`
	LinkSong string `json:"link_song"`
}

type Response struct {
	resp.Response
	//Status string `json:"status"`
	//Error  string `json:"error,omitempty"`
}

type SongSaver interface {
	SaveSong(groupToSave string, songToSave string, textSongToSave string, dateToSave string, linkToSave string) (int64, error)
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
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("failed to validate request", sl.Err(err))

			//render.JSON(w, r, resp.Error("invalid request"))
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		id, err := songSaver.SaveSong(req.Song, req.Group, req.TextSong, req.DateSong, req.LinkSong)

		if errors.Is(err, storage.ErrSongExist) {
			log.Info("song already exists", slog.String("song", req.Song))

			render.JSON(w, r, resp.Error("song already exists"))

			return
		}
		if err != nil {
			log.Error("failed to save song", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add song"))

			return
		}

		log.Info("song saved", slog.Int64("id", id))

		responseOK(w, r)

	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
