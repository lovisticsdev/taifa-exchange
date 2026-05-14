package exchange

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Route("/api/v1/exchange", func(r chi.Router) {
		r.Post("/authorize", handler.Authorize)
	})
}
