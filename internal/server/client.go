package server

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/mildlybrutal/websocketGo/internal/common"
)

type MyServerClient struct {
	*common.Client
}

func (c *MyServerClient) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c.Client
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)

	for {
		_, message, err := c.Conn.ReadMessage()

		if err != nil {
			break
		}

		var msg map[string]interface{}

		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg["type"] {
		case "broadcast":
			c.Hub.Broadcast <- message
		case "ping":
			pong, _ := json.Marshal(map[string]string{"type": "pong"})
			c.Send <- pong
		default:
			c.Send <- message
		}
	}
}

func (c *MyServerClient) WritePump() {
	defer c.Conn.Close()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
