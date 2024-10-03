package updateSong

import (
	resp "RestApi_v1/internal/config/internal/lib/api/response"
	"RestApi_v1/internal/config/internal/lib/logger/sl"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

type Request struct {
	ID       int    `json:"id"`
	Song     string `json:"song"`
	Group    string `json:"group,omitempty"`
	TextSong string `json:"text_song,omitempty"`
	DateSong string `json:"date_song,omitempty"`
	LinkSong string `json:"link_song,omitempty"`
}

type Response struct {
	resp.Response
}

type SongUpdater interface {
	UpdateSong(ID int, SongToSave string, GroupToSave string, TextSongToSave string, DateToSave string, LinkToSave string) (string, error)
}

func New(log *slog.Logger, songUpdater SongUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log = log.With(slog.String("op", "handlers.song.save.New"))

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.ValidationError(err.(validator.ValidationErrors)))
			return
		}

		_, err = songUpdater.UpdateSong(
			req.ID,
			req.Song,
			req.Group,
			req.TextSong,
			req.DateSong,
			req.LinkSong,
		)
		log.Info("Updating song with ID", slog.Int("ID", req.ID))

		if err != nil {

			log.Error("failed to update song", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to update song"))
			return
		}

		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
