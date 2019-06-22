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
	addr = flag.String("port", ":8010", "TCP port to listen to")
	cert = flag.String("cert", "cert.pem", "Path to certification")
	key  = flag.String("key", "key.pem", "Path to certification key")
)

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	flag.Parse()

	hub := socket.GetHub()
	go hub.Run()

	r := chi.NewRouter()
	r.Route("/realtime", func(r chi.Router) {
		r.Mount("/admin", routes.Endpoints(hub))
		r.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
			socket.ServeWs(hub, rw, r)
		})
	})

	srv := &http.Server{
		Addr:         *addr,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Println("")
		fmt.Println("Realtime server listening on ", *addr)
		fmt.Println("")
		if err := srv.ListenAndServeTLS(*cert, *key); err != nil {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	fmt.Println("\nShutting down websocket server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	fmt.Println("Realtime server stopped!")
}
