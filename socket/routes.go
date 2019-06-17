package socket

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// Routes return all realtime web socket admin routes
func Routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Logger,
	)

	// realtime/admin
	r.Route("/", func(r chi.Router) {
		r.Get("/services", getAllServices)
		r.Get("/clients", getAllClients)
	})

	return r
}

func getAllServices(w http.ResponseWriter, r *http.Request) {
	render.Status(r, 200)
	w.Write([]byte("welcome services"))
}

func getAllClients(w http.ResponseWriter, r *http.Request) {
	render.Status(r, 200)
	w.Write([]byte("welcome clients"))
}
