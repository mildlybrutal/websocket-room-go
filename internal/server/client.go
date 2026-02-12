package server

import (
	"context"
	"encoding/json"
	"html"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/mildlybrutal/websocketGo/internal/server/models"
	"golang.org/x/time/rate"
)

type MyServerClient struct {
	*common.Client
	limiter *rate.Limiter
}

func NewServerClient(baseClient *common.Client) *MyServerClient {
	return &MyServerClient{
		Client:  baseClient,
		limiter: rate.NewLimiter(rate.Limit(10), 20),
	}
}

func (c *MyServerClient) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c.Client
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

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

	if !c.limiter.Allow() {
		c.sendError("rate limit exceeded")
		log.Printf("Rate limit exceeded for client %s (user %d)", c.ID, c.UserID)
		return
	}

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
		c.handleRoomMessage(msg)
	case "private_message":
		if targetID, ok := msg["to"].(string); ok {
			c.sendPrivateMessage(targetID, message)
		}
	case "typing":
		c.handleTyping(msg)
	default:
		// Global broadcast
		c.Hub.Broadcast <- common.BroadcastMessage{Message: message, Sender: c.Client}

	}
}

func (c *MyServerClient) handleRoomMessage(msg map[string]any) {
	roomIDStr, ok := msg["room"].(string)
	if !ok || roomIDStr == "" {
		c.sendError("Room ID is required")
		return
	}

	content, ok := msg["content"].(string)
	if !ok {
		c.sendError("Message content is required")
		return
	}

	// Sanitize and validate content
	content = strings.TrimSpace(content)

	if content == "" {
		c.sendError("Message cannot be empty")
		return
	}

	if len(content) > 10000 {
		c.sendError("Message too long (max 10000 characters)")
		return
	}

	// HTML escape to prevent XSS
	content = html.EscapeString(content)

	// Validate room ID
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		c.sendError("Invalid room ID format")
		return
	}

	// Save to database
	chatEntry := &models.Chat{
		RoomID:   uint(roomID),
		SenderID: c.UserID,
		Content:  content,
	}

	if err := c.Hub.ChatRepo.SaveMessage(chatEntry); err != nil {
		log.Printf("DB Save Error for client %s: %v", c.ID, err)
		c.sendError("Failed to save message")
		return
	}

	// Prepare broadcast message
	broadcastMsg := map[string]any{
		"type":       "room_message",
		"room":       roomIDStr,
		"content":    content,
		"sender":     c.UserID,
		"sender_id":  c.ID,
		"message_id": chatEntry.ID,
		"timestamp":  chatEntry.CreatedAt.Unix(),
	}

	redisMsg, err := json.Marshal(broadcastMsg)
	if err != nil {
		log.Printf("JSON Marshal Error: %v", err)
		c.sendError("Failed to process message")
		return
	}

	// Publish to Redis (listener will broadcast to all servers)
	ctx := context.Background()
	channel := "room:" + roomIDStr

	err = c.Hub.RedisClient.Publish(ctx, channel, redisMsg).Err()
	if err != nil {
		log.Printf("Redis Publish Error for room %s: %v", roomIDStr, err)
		c.sendError("Failed to send message")
		return
	}

	log.Printf("Message published - user: %d, room: %s, msgID: %d", c.UserID, roomIDStr, chatEntry.ID)
}

func (c *MyServerClient) handleTyping(msg map[string]any) {
	roomID, ok := msg["room"].(string)

	if !ok || roomID == "" {
		return
	}

	typingMsg := map[string]any{
		"type":   "user_typing",
		"room":   roomID,
		"user":   c.UserID,
		"userID": c.ID,
	}

	redisMsg, _ := json.Marshal(typingMsg)

	ctx := context.Background()
	c.Hub.RedisClient.Publish(ctx, "room:"+roomID, redisMsg)
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
		select {
		case target.Send <- message:
			ack, _ := json.Marshal(map[string]any{
				"type":   "private_message_ack",
				"to":     targetID,
				"status": "delivered",
			})

			c.Send <- ack
		default:
			c.sendError("User is not available")
		}
	} else {
		c.sendError("User not found")
	}
}
