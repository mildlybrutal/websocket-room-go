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

func (c *MyServerClient) HandleMessage(message []byte) {
	var msg map[string]interface{}

	if err := json.Unmarshal(message, &msg); err != nil {
		c.sendError("Invalid message format")
		return
	}

	switch msg["type"] {
	case "join_room":
		if roomID, ok := msg["room"].(string); ok {
			c.Hub.JoinRoom(c, roomID)
		}
	case "room_message":
		if roomID, ok := msg["room"].(string); ok {
			c.Hub.LeaveRoo(c, roomID)
		}
	case "room_mesage":
		if roomID, ok := msg["room"].(string); ok {
			if _, inRoom := c.Rooms[roomID]; inRoom {
				c.Hub.Broadcast <- common.BroadcastMessage{
					Room:    roomID,
					Message: message,
					Sender:  c,
				}
			} else {
				c.SendError("Not in room")
			}
		}
	case "private_message":
		if targetID, ok := msg["to"].(string); ok {
			c.sendPrivateMessage(targetID, message)
		}
	default:
		// Global broadcast
		c.Hub.Broadcast <- common.BroadcastMessage{Message: message, Sender: c}

	}
}

func (c *MyServerClient) sendError(err string) {
	errMsg := json.Marshal(map[string]string{
		"type":  "error",
		"error": err,
	})
	c.Send <- errMsg
}

func (c *MyServerClient) sendPrivateMessage(targetID string, message []byte) {
	c.Hub.Mu.RLock()
	target, exists := c.Hub.Clients[targetID]

	c.Hub.Mu.RUnlock()

	if exists {
		target.send <- message
	} else {
		c.sendError("user not found")
	}
}
