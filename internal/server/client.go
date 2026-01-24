package server

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/mildlybrutal/websocketGo/internal/server/models"
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.HandleMessage(message)
	}
}

func (c *MyServerClient) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}

	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *MyServerClient) HandleMessage(message []byte) {
	var msg map[string]any

	if err := json.Unmarshal(message, &msg); err != nil {
		c.sendError("Invalid message format")
		return
	}

	msgType, _ := msg["type"].(string)

	switch msgType {
	case "join_room":
		if roomID, ok := msg["room"].(string); ok {
			c.Hub.JoinRoom(roomID, c.Client)
		}
	case "leave_room":
		if roomID, ok := msg["room"].(string); ok {
			c.Hub.LeaveRoom(c.Client, roomID)
		}
	case "room_message":
		if roomIDStr, ok := msg["room"].(string); ok {
			roomID, _ := strconv.ParseUint(roomIDStr, 10, 32)
			content, _ := msg["content"].(string)

			// Persist to DB
			chatEntry := &models.Chat{
				RoomID:   uint(roomID),
				SenderID: 1,
				Content:  content,
			}

			if err := c.Hub.ChatRepo.SaveMessage(chatEntry); err != nil {
				log.Printf("DB Save Error: %v", err)
			}

			// Broadcast
			c.Hub.Broadcast <- common.BroadcastMessage{
				Room:    roomIDStr,
				Message: message,
				Sender:  c.Client,
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
