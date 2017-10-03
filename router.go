package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	wr "github.com/stuwilli/go-web-response"
)

//SetupRouter ...
func SetupRouter() *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		resp := wr.NewBuilder().Status(http.StatusNotFound).Build()
		resp.WriteJSON(w)
	}))

	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		resp := wr.NewBuilder().Status(http.StatusMethodNotAllowed).Build()
		resp.WriteJSON(w)
	}))

	return r
}
