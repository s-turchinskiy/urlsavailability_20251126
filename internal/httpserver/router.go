package httpserver

import (
	"github.com/go-chi/chi/v5"
)

func Router(h *HTTPServerHandlers) *chi.Mux {

	router := chi.NewRouter()
	router.Route("/statuses", func(r chi.Router) {
		r.Post("/", h.Statuses)
	})

	router.Route("/pdf", func(r chi.Router) {
		r.Post("/", h.PDF)
	})

	return router

}
