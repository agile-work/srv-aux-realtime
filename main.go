package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/agile-work/srv-aux-realtime/socket"
)

var addr = flag.String("addr", ":8080", "realtime service port")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./client/client.html")
}

func main() {
	flag.Parse()
	hub := socket.NewHub()
	go hub.Run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket.ServeWs(hub, w, r)
	})
	err := http.ListenAndServeTLS(*addr, "cert.pem", "key.pem", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
