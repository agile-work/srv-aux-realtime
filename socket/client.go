package socket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub         *Hub
	id          string
	scope       string
	connections map[*Connection]bool
	register    chan *Connection
	unregister  chan *Connection
	inbox       chan *Message
	outbox      chan *Message
}

// Close all channels
func (c *Client) Close() {
	close(c.register)
	close(c.unregister)
	close(c.inbox)
	close(c.outbox)
}

// NewClient creates a new client
func NewClient(hub *Hub, id, scope string, conn *Connection) *Client {
	client := &Client{
		hub:         hub,
		id:          id,
		scope:       scope,
		connections: make(map[*Connection]bool),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		inbox:       make(chan *Message, 256),
		outbox:      make(chan *Message, 256),
	}

	client.connections[conn] = true

	return client
}

// Run start a new thread to process hub actions
func (c *Client) Run() {
	for {
		select {
		case conn := <-c.register:
			c.connections[conn] = true
			go conn.writePump()
			go conn.readPump()
		case conn := <-c.unregister:
			if _, ok := c.connections[conn]; ok {
				delete(c.connections, conn)
				close(conn.send)
			}
		case message := <-c.outbox:
			c.hub.messages <- message
		case message := <-c.inbox:
			for conn := range c.connections {
				msgBytes, err := message.GetData()
				if err != nil {
					log.Println("invalid message data")
				} else {
					conn.send <- msgBytes
				}
			}

		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	scope := "user"
	if token == "" {
		token = r.Header.Get("Authorization")
		scope = "service"
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	id, scope, err := parseToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	connection := &Connection{
		conn: conn,
		send: make(chan []byte, 256),
	}

	client := hub.GetClient(id)
	if client == nil {
		client := NewClient(hub, id, scope, connection)
		client.hub.register <- client
	}

	connection.client = client
	client.register <- connection
}

func parseToken(token string) (string, string, error) {
	return token, "user", nil
}
