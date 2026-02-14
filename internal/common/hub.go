package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/mildlybrutal/websocketGo/internal/server/models"
	"github.com/redis/go-redis/v9"
)

type Hub struct {
	Clients    map[string]*Client
	Rooms      map[string]*Room
	Broadcast  chan BroadcastMessage
	Register   chan *Client
	Unregister chan *Client

	RedisClient         *redis.Client
	activeSubscriptions map[string]context.CancelFunc

	ctx    context.Context
	cancel context.CancelFunc

	Mu sync.RWMutex

	ChatRepo interface {
		SaveMessage(chat *models.Chat) error
		GetRoomHistory(roomID uint, limit int) ([]models.Chat, error)
	}
}

func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		Clients:             make(map[string]*Client),
		Rooms:               make(map[string]*Room),
		Broadcast:           make(chan BroadcastMessage),
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		activeSubscriptions: make(map[string]context.CancelFunc),
		ctx:                 ctx,
		cancel:              cancel,
	}
}

func (h *Hub) Run() {
	//Periodic cleanup ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		// select lets the hub listen to multiple channels at once.
		select {
		case client := <-h.Register:
			//New user connected
			h.Mu.Lock() // exclusive for the client
			h.Clients[client.ID] = client
			h.Mu.Unlock()
			log.Printf("Client %s registered", client.ID)

		case client := <-h.Unregister:
			//user disconnection
			h.Mu.Lock()
			if _, ok := h.Clients[client.ID]; ok {
				for roomID := range h.Rooms {
					if room, exists := h.Rooms[roomID]; exists {
						room.RemoveClient(client)
						if room.IsEmpty() {
							if cancelFunc, exists := h.activeSubscriptions[roomID]; exists {
								cancelFunc()
								delete(h.activeSubscriptions, roomID)
							}
							delete(h.Rooms, roomID)
						}
					}
				}
				delete(h.Clients, client.ID) //remove from list
				close(client.Send)           //shuts down client's individual channel
				log.Printf("Client %s unregistered", client.ID)
			}
			h.Mu.Unlock() // Unlock before broadcasting
		case message := <-h.Broadcast:
			// A message was received from one client that needs to go to everyone.
			h.broadcastMessage(message)
		case <-ticker.C:
			h.cleanup()
		}
	}

}

func (h *Hub) Shutdown() {
	log.Println("Initiating hub shutdown...")

	h.Mu.Lock()
	for roomID, cancelFunc := range h.activeSubscriptions {
		log.Printf("Stopping Redis listener for room: %s", roomID)
		cancelFunc()
	}
	h.Mu.Unlock()

	h.cancel()
	log.Println("Hub shutdown complete")
}

func (h *Hub) ListenToRedis(roomID string) {
	// Create cancellable context for this subscription
	ctx, cancel := context.WithCancel(h.ctx)

	// Register cancel function for later cleanup
	h.Mu.Lock()
	h.activeSubscriptions[roomID] = cancel
	h.Mu.Unlock()

	// Ensure cleanup on function exit
	defer func() {
		h.Mu.Lock()
		delete(h.activeSubscriptions, roomID)
		h.Mu.Unlock()

		// Recover from any panics
		if r := recover(); r != nil {
			log.Printf("ERROR: Panic in Redis listener for room %s: %v", roomID, r)
		}

		log.Printf("Redis listener stopped for room: %s", roomID)
	}()

	channel := "room:" + roomID

	// Subscribe to Redis channel
	pubsub := h.RedisClient.Subscribe(ctx, channel)
	defer pubsub.Close()

	// Wait for subscription confirmation with timeout
	subCtx, subCancel := context.WithTimeout(ctx, 5*time.Second)
	if _, err := pubsub.Receive(subCtx); err != nil {
		subCancel()
		log.Printf("ERROR: Failed to subscribe to Redis channel %s: %v", channel, err)
		return
	}
	subCancel()

	log.Printf("Started Redis listener for room: %s", roomID)

	// Get message channel
	msgChan := pubsub.Channel()

	// Message processing loop
	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown or room cleanup
			return

		case msg, ok := <-msgChan:
			// Channel closed
			if !ok {
				log.Printf("Redis channel closed for room: %s", roomID)
				return
			}

			// Nil message - skip
			if msg == nil {
				continue
			}

			// Validate payload
			if len(msg.Payload) == 0 {
				log.Printf("Warning: Empty payload for room %s", roomID)
				continue
			}

			// Sanity check - prevent extremely large messages
			if len(msg.Payload) > 10*1024*1024 { // 10MB
				log.Printf("Warning: Message too large (%d bytes) for room %s",
					len(msg.Payload), roomID)
				continue
			}

			// Broadcast to all local clients in this room
			h.broadcastToLocalClients(roomID, []byte(msg.Payload))
		}
	}
}

func (h *Hub) broadcastToLocalClients(roomID string, message []byte) {
	h.Mu.RLock()
	room, exists := h.Rooms[roomID]
	h.Mu.RUnlock()

	// Room doesn't exist locally - this is normal in multi-server setup
	if !exists {
		return
	}

	room.Mu.RLock()
	defer room.Mu.RUnlock()

	// Send to all clients in the room
	successCount := 0
	for client := range room.Clients {
		select {
		case client.Send <- message:
			successCount++
		default:
			// Buffer full - log but don't block
			log.Printf("Warning: Client %s buffer full (room: %s)", client.ID, roomID)
		}
	}

	// Log broadcast statistics
	if successCount > 0 {
		log.Printf("Broadcasted to %d clients in room %s", successCount, roomID)
	}
}

func (h *Hub) broadcastMessage(message BroadcastMessage) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	if message.Room != "" {
		if room, exists := h.Rooms[message.Room]; exists {
			room.Broadcast(message.Message, message.Sender)
		} else {
			for _, client := range h.Clients {
				if client != message.Sender {
					select {
					case client.Send <- message.Message:
					default:
						//handle full buffer
					}
				}
			}
		}
	}
}

func (h *Hub) cleanup() {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	for id, room := range h.Rooms {
		if room.IsEmpty() {
			delete(h.Rooms, id)
		}
	}
}

func (h *Hub) JoinRoom(RoomID string, client *Client) error {
	h.Mu.Lock()
	room, exists := h.Rooms[RoomID]
	if !exists {
		room = &Room{
			ID:      RoomID,
			Clients: make(map[*Client]bool),
		}
		h.Rooms[RoomID] = room
	}

	// Start Redis listener if not already active
	if _, subscribed := h.activeSubscriptions[RoomID]; !subscribed {
		go h.ListenToRedis(RoomID)
	}

	h.Mu.Unlock()

	room.AddClient(client)

	client.Mu.Lock()
	if client.Rooms == nil {
		client.Rooms = make(map[string]bool)
	}
	client.Rooms[RoomID] = true
	client.Mu.Unlock()

	// Load chat history
	id, _ := strconv.ParseUint(RoomID, 10, 32)
	history, err := h.ChatRepo.GetRoomHistory(uint(id), 50)

	if err == nil {
		for _, msg := range history {
			historyJSON, _ := json.Marshal(map[string]any{
				"type":        "history_message",
				"content":     msg.Content,
				"sender":      msg.SenderID,
				"sender_name": msg.Sender.Username,
				"time":        msg.CreatedAt,
			})
			client.Send <- historyJSON
		}
	}

	// Notify others via Redis
	notification, _ := json.Marshal(map[string]any{
		"type":   "user_joined_room",
		"room":   RoomID,
		"userId": client.ID,
	})

	ctx := context.Background()
	h.RedisClient.Publish(ctx, "room:"+RoomID, notification)

	// Send room info to joining client
	roomInfo, _ := json.Marshal(map[string]any{
		"type":    "room_joined",
		"room":    RoomID,
		"members": room.GetMemberIDs(),
	})
	client.Send <- roomInfo

	return nil
}

func (h *Hub) LeaveRoom(client *Client, RoomID string) {
	h.Mu.Lock()
	room, exists := h.Rooms[RoomID]
	if !exists {
		h.Mu.Unlock()
		return
	}
	h.Mu.Unlock()

	room.RemoveClient(client)

	client.Mu.Lock()
	delete(client.Rooms, RoomID)
	client.Mu.Unlock()

	notification, _ := json.Marshal(map[string]any{
		"type":   "user_left_room",
		"room":   RoomID,
		"userID": client.ID,
	})

	ctx := context.Background()
	h.RedisClient.Publish(ctx, "room:"+RoomID, notification)

	h.Mu.Lock()
	if room.IsEmpty() {
		if cancelFunc, exists := h.activeSubscriptions[RoomID]; exists {
			cancelFunc()
			delete(h.activeSubscriptions, RoomID)
		}
		delete(h.Rooms, RoomID)
	}
	h.Mu.Unlock()
}

func (h *Hub) SetUserOnline(userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("online: %d", userID)
	return h.RedisClient.Set(ctx, key, "1", 30*time.Second).Err()
}

func (h *Hub) SetUserOffline(userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("online: %d", userID)
	return h.RedisClient.Del(ctx, key).Err()
}

func (h *Hub) IsUserOnline(userID uint) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("online:%d", userID)
	result, err := h.RedisClient.Exists(ctx, key).Result()
	return result > 0, err
}

func (h *Hub) GetOnlineUsers() ([]uint, error) {
	ctx := context.Background()
	keys, err := h.RedisClient.Keys(ctx, "online:*").Result()
	if err != nil {
		return nil, err
	}

	onlineUsers := make([]uint, 0, len(keys))
	for _, key := range keys {
		var userID uint
		if _, err := fmt.Sscanf(key, "online:%d", &userID); err == nil {
			onlineUsers = append(onlineUsers, userID)
		}
	}
	return onlineUsers, nil
}

func (h *Hub) StartHeartbeat(client *Client) {
	if client.UserID == 0 {
		return
	}

	ticker := time.NewTicker(10 * time.Second)

	go func() {
		defer ticker.Stop()

		h.SetUserOnline(client.UserID)

		for {
			select {
			case <-ticker.C:
				if err := h.SetUserOnline(client.UserID); err != nil {
					log.Printf("Failed to refresh online status for user %d: %v", client.UserID, err)
					return
				}
			case <-h.ctx.Done():
				h.SetUserOffline(client.UserID)
				return
			}
		}
	}()
}
