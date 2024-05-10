package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/muhafs/orders-api/handler"
)

func loadRoutes() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", loadOrderRoutes)

	return router
}

func loadOrderRoutes(router chi.Router) {
	orderHandler := new(handler.Order)

	router.Post("/", orderHandler.Create)
	router.Get("/", orderHandler.List)
	router.Get("/{id}", orderHandler.Find)
	router.Put("/{id}", orderHandler.Update)
	router.Delete("/{id}", orderHandler.Delete)
}
