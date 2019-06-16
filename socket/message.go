package socket

import "encoding/json"

// Message defines the message that travels through the websocket
type Message struct {
	Recipients []string    `json:"recipients"`
	Data       interface{} `json:"data"`
}

// NewMessage creates a new message
func NewMessage(data interface{}, recipients ...string) *Message {
	return &Message{
		Data:       data,
		Recipients: recipients,
	}
}

// GetData returns message data byte array
func (m *Message) GetData() ([]byte, error) {
	return json.Marshal(m.Data)
}
