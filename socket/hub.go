package socket

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan string

	// Messages from the web socket
	messages chan *Message
}

// NewHub initialize a new hub
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan string),
		clients:    make(map[string]*Client),
		messages:   make(chan *Message, 256),
	}
}

// Run start a new thread to process hub actions
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.id] = client
			go client.Run()
		case clientID := <-h.unregister:
			if client, ok := h.clients[clientID]; ok {
				delete(h.clients, clientID)
				client.Close()
			}
		case message := <-h.messages:
			if len(message.Recipients) <= len(h.clients) {
				for _, id := range message.Recipients {
					if client, ok := h.clients[id]; ok {
						client.inbox <- message
					}
				}
			} else {
				for _, client := range h.clients {
					if contains(message.Recipients, client.id) {
						client.inbox <- message
					}
				}
			}
		}
	}
}

// GetClient returns a client based on the id
func (h *Hub) GetClient(id string) *Client {
	if client, ok := h.clients[id]; ok {
		return client
	}
	return nil
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
