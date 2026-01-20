package server

import (
	"encoding/json"
	"log"

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
			c.Hub.Broadcast <- common.BroadcastMessage{}
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
			c.Hub.JoinRoom(roomID, c.Client)
		}
	case "leave_room":
		if roomID, ok := msg["room"].(string); ok {
			c.Hub.LeaveRoom(c.Client, roomID)
		}
	case "room_message":
		if roomID, ok := msg["room"].(string); ok {
			if _, inRoom := c.Rooms[roomID]; inRoom {
				c.Hub.Broadcast <- common.BroadcastMessage{
					Room:    roomID,
					Message: message,
					Sender:  c.Client,
				}
			} else {
				c.sendError("Not in room")
			}
		}
	case "private_message":
		if targetID, ok := msg["to"].(string); ok {
			c.sendPrivateMessage(targetID, message)
		}
	default:
		// Global broadcast
		c.Hub.Broadcast <- common.BroadcastMessage{Message: message, Sender: c.Client}

	}
}

func (c *MyServerClient) sendError(errStr string) {
	// json.Marshal returns ([]byte, error)
	errMsg, err := json.Marshal(map[string]string{
		"type":  "error",
		"error": errStr,
	})

	// Check if marshaling itself failed
	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		return
	}

	// Now errMsg is of type []byte and can be sent to the channel
	c.Send <- errMsg
}
func (c *MyServerClient) sendPrivateMessage(targetID string, message []byte) {
	c.Hub.Mu.RLock()
	target, exists := c.Hub.Clients[targetID]

	c.Hub.Mu.RUnlock()

	if exists {
		target.Send <- message
	} else {
		c.sendError("user not found")
	}
}
