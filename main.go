package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	config := ServiceConfig{
		Backend: "http://localhost:8080/backend",
		Path:    "/test/*",
		Methods: []string{"GET"},
		Auth: AuthConfig{
			UseACM:       true,
			RequiredRole: []float64{0, 1},
		},
	}

	config.CreateProxy(r)

	r.HandleFunc("/backend*", func(w http.ResponseWriter, r *http.Request) {

		str, _ := httputil.DumpRequest(r, false)
		fmt.Printf("%s", str)
		w.WriteHeader(200)
		w.Write(([]byte("CHEESE")))
	})

	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		str, _ := httputil.DumpRequest(r, false)
		fmt.Printf("%s", str)
		w.WriteHeader(404)
		w.Write(([]byte("404 Not found")))
	}))

	http.ListenAndServe(":8080", r)
}
