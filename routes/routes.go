package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/agile-work/srv-shared/service"
	"github.com/agile-work/srv-shared/socket"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// Endpoints return all realtime web socket admin routes
func Endpoints(hub *socket.Hub, aux *service.Service) *chi.Mux {
	r := chi.NewRouter()

	r.Use(
		middleware.Logger,
	)

	// realtime/admin
	r.Route("/", func(r chi.Router) {
		r.Get("/clients", func(w http.ResponseWriter, r *http.Request) {
			getClients(w, r, hub, aux)
		})
	})

	return r
}

func getClients(w http.ResponseWriter, r *http.Request, hub *socket.Hub, aux *service.Service) {
	clients := make(map[string]interface{})

	totalUsers := 0
	totalServices := 0
	totalConnections := 0

	services := []json.RawMessage{}
	totalServices++
	aux.GetUptime()
	auxBytes, _ := json.Marshal(aux)
	services = append(services, auxBytes)

	users := []json.RawMessage{}
	for _, c := range hub.GetClients() {
		if c.GetScope() == "user" {
			totalUsers++
			usr, _ := c.GetUserData()
			users = append(users, usr)
		} else if c.GetScope() == "service" {
			totalServices++
			srv, _ := c.GetServiceData()
			services = append(services, srv)
		}
		totalConnections += c.GetTotalConnections()
	}

	diff := time.Now().Sub(hub.GetStartAt())
	uptime := fmt.Sprint(diff)

	clients["metrics"] = map[string]interface{}{
		"total":       len(hub.GetClients()),
		"connections": totalConnections,
		"users":       totalUsers,
		"services":    totalServices,
		"start_at":    hub.GetStartAt(),
		"uptime":      uptime,
	}
	clients["services"] = services
	clients["users"] = users

	render.Status(r, 200)
	render.JSON(w, r, clients)
}
