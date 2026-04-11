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
	"github.com/mildlybrutal/websocketGo/internal/metrics"
	"github.com/mildlybrutal/websocketGo/internal/repository"
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

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

func (c *MyServerClient) ReadPump() {
	defer func() {
		metrics.ActiveConnections.Dec()
		c.Hub.Unregister <- c.Client
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

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
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
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

	case "edit_message":
		c.handleEditMessage(msg)
	case "delete_message":
		c.handleDeleteMessage(msg)
	case "mark_read":
		c.handleMarkRead(msg)
	case "get_online_users":
		c.handlePresenceRequest(msg)
	default:
		// Global broadcast
		c.Hub.Broadcast <- common.BroadcastMessage{Message: message, Sender: c.Client}

	}
}

func (c *MyServerClient) handleRoomMessage(msg map[string]any) {
	start := time.Now()
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
	content = strings.TrimSpace(content)
	if content == "" {
		c.sendError("Message cannot be empty")
		return
	}
	if len(content) > 10000 {
		c.sendError("Message too long (max 10000 characters)")
		return
	}
	content = html.EscapeString(content)
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		c.sendError("Invalid room ID format")
		return
	}

	// Declare outside so broadcastMsg can access it
	var messageID uint
	var timestamp int64

	if c.UserID != 0 {
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
		messageID = chatEntry.ID
		timestamp = chatEntry.CreatedAt.Unix()
	} else {
		timestamp = time.Now().Unix()
	}

	broadcastMsg := map[string]any{
		"type":        "room_message",
		"room":        roomIDStr,
		"content":     content,
		"sender":      c.UserID,
		"sender_id":   c.ID,
		"sender_name": c.Username,
		"message_id":  messageID,
		"timestamp":   timestamp,
	}

	redisMsg, err := json.Marshal(broadcastMsg)
	if err != nil {
		log.Printf("JSON Marshal Error: %v", err)
		c.sendError("Failed to process message")
		return
	}

	ctx := context.Background()
	channel := "room:" + roomIDStr

	err = c.Hub.RedisClient.Publish(ctx, channel, redisMsg).Err()
	if err != nil {
		metrics.RedisPublishErrors.Inc()
		log.Printf("Redis Publish Error for room %s: %v", roomIDStr, err)
		c.sendError("Failed to send message")
		return
	}

	metrics.MessagesTotal.WithLabelValues(roomIDStr).Inc()
	metrics.MessageLatency.Observe(time.Since(start).Seconds())

	log.Printf("Message published - user: %d, room: %s, msgID: %d", c.UserID, roomIDStr, messageID)
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
	target, localExists := c.Hub.Clients[targetID]
	c.Hub.Mu.RUnlock()

	if localExists {
		select {
		case target.Send <- message:
			ack, _ := json.Marshal(map[string]any{
				"type":   "private_message_ack",
				"to":     targetID,
				"status": "delivered",
			})
			c.Send <- ack
			return
		default:
			c.sendError("User is busy (buffer full)")
			return
		}
	}

	ctx := context.Background()

	err := c.Hub.RedisClient.Publish(ctx, "user:"+targetID, message).Err()

	if err != nil {
		c.sendError("Failed to send message")
		return
	}

	ack, _ := json.Marshal(map[string]any{
		"type":   "private_message_ack",
		"to":     targetID,
		"status": "sent",
	})
	c.Send <- ack
}

func (c *MyServerClient) handleEditMessage(msg map[string]any) {
	messageIDFloat, ok := msg["message_id"].(float64)
	if !ok {
		c.sendError("message_id is required")
		return
	}
	messageID := uint(messageIDFloat)

	newContent, ok := msg["content"].(string)
	if !ok || strings.TrimSpace(newContent) == "" {
		c.sendError("content is required")
		return
	}

	roomIDStr, ok := msg["room"].(string)
	if !ok {
		c.sendError("room is required")
		return
	}

	// Sanitize content
	newContent = strings.TrimSpace(html.EscapeString(newContent))

	// Update in database
	repo, ok := c.Hub.ChatRepo.(*repository.ChatRepository)
	if !ok {
		c.sendError("Internal error")
		return
	}

	if err := repo.EditMessage(messageID, newContent, c.UserID); err != nil {
		log.Printf("Failed to edit message %d: %v", messageID, err)
		c.sendError("Failed to edit message")
		return
	}

	// Broadcast edit notification via Redis
	editNotif := map[string]any{
		"type":        "message_edited",
		"room":        roomIDStr,
		"message_id":  messageID,
		"content":     newContent,
		"edited_by":   c.UserID,
		"editor_name": c.Username,
		"edited_at":   time.Now().Unix(),
	}

	redisMsg, _ := json.Marshal(editNotif)
	ctx := context.Background()
	c.Hub.RedisClient.Publish(ctx, "room:"+roomIDStr, redisMsg)

	log.Printf("Message %d edited by user %d in room %s", messageID, c.UserID, roomIDStr)
}

func (c *MyServerClient) handleDeleteMessage(msg map[string]any) {
	messageIDFloat, ok := msg["message_id"].(float64)
	if !ok {
		c.sendError("message_id is required")
		return
	}
	messageID := uint(messageIDFloat)

	roomIDStr, ok := msg["room"].(string)
	if !ok {
		c.sendError("room is required")
		return
	}

	// Delete in database
	repo, ok := c.Hub.ChatRepo.(*repository.ChatRepository)
	if !ok {
		c.sendError("Internal error")
		return
	}

	if err := repo.DeleteMessage(messageID, c.UserID); err != nil {
		log.Printf("Failed to delete message %d: %v", messageID, err)
		c.sendError("Failed to delete message")
		return
	}

	// Broadcast delete notification via Redis
	deleteNotif := map[string]any{
		"type":         "message_deleted",
		"room":         roomIDStr,
		"message_id":   messageID,
		"deleted_by":   c.UserID,
		"deleter_name": c.Username,
		"deleted_at":   time.Now().Unix(),
	}

	redisMsg, _ := json.Marshal(deleteNotif)
	ctx := context.Background()
	c.Hub.RedisClient.Publish(ctx, "room:"+roomIDStr, redisMsg)

	log.Printf("Message %d deleted by user %d in room %s", messageID, c.UserID, roomIDStr)
}

func (c *MyServerClient) handleMarkRead(msg map[string]any) {
	messageIDsRaw, ok := msg["message_ids"].([]interface{})
	if !ok {
		c.sendError("message_ids array is required")
		return
	}

	repo, ok := c.Hub.ChatRepo.(*repository.ChatRepository)
	if !ok {
		c.sendError("Internal error")
		return
	}

	readCount := 0
	for _, idRaw := range messageIDsRaw {
		if idFloat, ok := idRaw.(float64); ok {
			messageID := uint(idFloat)
			if err := repo.MarkMessageAsRead(messageID, c.UserID); err == nil {
				readCount++
			}
		}
	}

	// Send acknowledgment
	ack, _ := json.Marshal(map[string]any{
		"type":         "read_receipt_ack",
		"marked_count": readCount,
		"user_id":      c.UserID,
	})
	c.Send <- ack

	log.Printf("User %d marked %d messages as read", c.UserID, readCount)
}

func (c *MyServerClient) handlePresenceRequest(msg map[string]any) {
	onlineUsers, err := c.Hub.GetOnlineUsers()
	if err != nil {
		log.Printf("Failed to get online users: %v", err)
		c.sendError("Failed to fetch online users")
		return
	}

	response, _ := json.Marshal(map[string]any{
		"type":         "presence_update",
		"online_users": onlineUsers,
	})

	c.Send <- response
}
