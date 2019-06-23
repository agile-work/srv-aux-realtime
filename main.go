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
	fmt.Println("Starting Service Realtime...")

	hub := socket.GetHub()
	go hub.Run()

	r := chi.NewRouter()
	r.Route("/realtime", func(r chi.Router) {
		r.Mount("/admin", routes.Endpoints(hub))
		r.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
			socket.ServeWs(hub, rw, r)
		})
	})

	addr := fmt.Sprintf("%s:%d", *host, *port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("Realtime pid:%d listening on %d\n", os.Getpid(), *port)
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
