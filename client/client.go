package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/agile-work/srv-aux-realtime/socket"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var id = flag.String("id", "", "Client id")
var scope = flag.String("scope", "user", "Client scope")
var message = flag.String("msg", "", "Message")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	// TODO: remove the InsecureSkipVerify when deploy in production
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	dialer := websocket.Dialer{
		TLSClientConfig: config,
	}

	token := *id + "," + *scope
	c, resp, err := dialer.Dial(u.String(), http.Header{"Authorization": []string{token}})

	if err != nil {
		if err == websocket.ErrBadHandshake {
			log.Printf("handshake failed with status %d", resp.StatusCode)
		}
		log.Fatal(err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recived message: %s", message)
		}
	}()

	if *scope == "service" {
		msg := socket.Message{
			Recipients: []string{"000001", "000002"},
			Data:       *message,
		}

		jsonBytes, _ := json.Marshal(msg)

		err := c.WriteMessage(websocket.TextMessage, jsonBytes)
		if err != nil {
			log.Println("write:", err)
			return
		}
		log.Println("Message sent")
	}

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
