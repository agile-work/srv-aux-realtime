package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/agile-work/srv-aux-realtime/socket"
)

var addr = flag.String("addr", ":8080", "realtime service port")

func main() {
	flag.Parse()
	hub := socket.NewHub()
	go hub.Run()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
