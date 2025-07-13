package delete

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.4 --name=URLDeletter
type URLDeletter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeletter URLDeletter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		err := urlDeletter.DeleteURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.JSON(w, r, resp.Error("not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", "error", err)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("url deleted")

		render.JSON(w, r, fmt.Sprintf("url by alias: %s deleted", alias))
	}
}
