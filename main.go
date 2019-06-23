package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/agile-work/srv-aux-realtime/routes"
	"github.com/agile-work/srv-shared/constants"
	"github.com/agile-work/srv-shared/service"
	"github.com/agile-work/srv-shared/socket"
	"github.com/go-chi/chi"
)

var (
	cert = flag.String("cert", "cert.pem", "Path to certification")
	key  = flag.String("key", "key.pem", "Path to certification key")
	host = flag.String("host", "localhost", "Realtime host")
	port = flag.Int("port", 8010, "Realtime port")
)

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	flag.Parse()
	pid := os.Getpid()
	aux := service.New("Realtime", constants.ServiceTypeAuxiliary, *host, *port, pid)

	fmt.Printf("Starting Service %s...\n", aux.Name)
	fmt.Printf("[Instance: %s | PID: %d]\n", aux.InstanceCode, aux.PID)

	hub := socket.GetHub()
	go hub.Run()

	r := chi.NewRouter()
	r.Route("/realtime", func(r chi.Router) {
		r.Mount("/admin", routes.Endpoints(hub, aux))
		r.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
			socket.ServeWs(hub, rw, r)
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", aux.Host, aux.Port),
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("Realtime ready listening on %d\n", aux.Port)
		if err := srv.ListenAndServeTLS(*cert, *key); err != nil {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	fmt.Println("\nShutting down realtime server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	fmt.Println("Realtime server stopped!")
}
