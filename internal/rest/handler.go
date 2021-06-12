package rest

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/cache"
	"github.com/soupedup/purgery/internal/common"
	"github.com/soupedup/purgery/internal/rest/internal/middleware"
	"github.com/soupedup/purgery/internal/rest/internal/render"
)

func newHandler(logger *zap.Logger, cache *cache.Cache, apiKey string) http.Handler {
	r := &handler{
		Router: new(httprouter.Router),
		logger: logger,
		cache:  cache,
	}

	r.HandlerFunc(http.MethodGet, "/health", r.health)

	purge := http.HandlerFunc(r.purge)
	r.Handler(http.MethodPost, "/purge", middleware.Auth(apiKey, purge))

	return middleware.Log(logger, r)
}

type handler struct {
	*httprouter.Router
	logger *zap.Logger
	cache  *cache.Cache
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	if h.cache.Ping(r.Context(), h.logger) {
		render.NoContent(w)
	} else {
		render.ServiceUnavailable(w)
	}
}

func (h *handler) purge(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		URL string `json:"url"`
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&payload); err != nil || !common.IsValidURL(payload.URL) {
		render.UnprocessableEntity(w)

		return
	}

	if err := h.cache.Purge(r.Context(), h.logger, payload.URL); err != nil {
		render.InternalServerError(w)

		return
	}

	render.NoContent(w)
}
