package routes

import (
	"net/http"

	"github.com/agile-work/srv-shared/socket"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// Endpoints return all realtime web socket admin routes
func Endpoints(hub *socket.Hub) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Logger,
	)

	// realtime/admin
	r.Route("/", func(r chi.Router) {
		r.Get("/clients", func(w http.ResponseWriter, r *http.Request) {
			getClients(w, r, hub)
		})
	})

	return r
}

func getClients(w http.ResponseWriter, r *http.Request, hub *socket.Hub) {
	clients := make(map[string]interface{})
	clients["total"] = len(hub.GetClients())
	totalClients := 0
	totalConnections := 0
	services := []string{}
	for _, c := range hub.GetClients() {
		if c.GetScope() == "user" {
			totalClients++
		} else if c.GetScope() == "service" {
			services = append(services, c.GetID())
		}
		totalConnections += c.GetTotalConnections()
	}
	clients["users"] = totalClients
	clients["connections"] = totalConnections
	clients["services"] = services

	render.Status(r, 200)
	render.JSON(w, r, clients)
}
