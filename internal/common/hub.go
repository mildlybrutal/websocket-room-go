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
	activeSubscriptions map[string]bool

	Mu sync.RWMutex

	ChatRepo interface {
		SaveMessage(chat *models.Chat) error
		GetRoomHistory(roomID uint, limit int) ([]models.Chat, error)
	}
}

func NewHub() *Hub {
	return &Hub{
		Clients:             make(map[string]*Client),
		Rooms:               make(map[string]*Room),
		Broadcast:           make(chan BroadcastMessage),
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		activeSubscriptions: make(map[string]bool),
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

func (h *Hub) ListenToRedis(ctx context.Context, roomID string) {
	pubsub := h.RedisClient.Subscribe(ctx, "room:"+roomID)
	defer pubsub.Close()

	defer func() {
		h.Mu.Lock()
		delete(h.activeSubscriptions, roomID)
		h.Mu.Unlock()
	}()

	ch := pubsub.Channel()

	// This loop runs until the channel is closed or context is cancelled
	for msg := range ch {
		// We broadcast to LOCAL clients only
		// The message originated from Redis (potentially another server)
		h.broadcastToLocalClients(roomID, []byte(msg.Payload))
	}

	// Cleanup when loop exits
	h.Mu.Lock()
	delete(h.activeSubscriptions, roomID)
	h.Mu.Unlock()
}

func (h *Hub) broadcastToLocalClients(roomID string, message []byte) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	if room, ok := h.Rooms[roomID]; ok {
		for client := range room.Clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(room.Clients, client)
			}
		}
	}
}

func (h *Hub) handleIncomingMessage(msg Message) {
	ctx := context.Background()

	payload, _ := json.Marshal(msg)

	channel := fmt.Sprintf("room: %s", msg.RoomID)

	err := h.RedisClient.Publish(ctx, channel, payload).Err()
	if err != nil {
		log.Printf("Failed to publish to Redis: %v", err)
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

	if !h.activeSubscriptions[RoomID] {
		go h.ListenToRedis(context.Background(), RoomID)
		h.activeSubscriptions[RoomID] = true
	}

	h.Mu.Unlock()

	room.AddClient(client)

	client.Mu.Lock()

	if client.Rooms == nil {
		client.Rooms = make(map[string]bool)
	}

	client.Rooms[RoomID] = true

	client.Mu.Unlock()

	id, _ := strconv.ParseUint(RoomID, 10, 32)

	history, err := h.ChatRepo.GetRoomHistory(uint(id), 50)

	if err == nil {
		for _, msg := range history {
			historyJSON, _ := json.Marshal(map[string]any{
				"type":    "history_message",
				"content": msg.Content,
				"sender":  msg.SenderID,
				"time":    msg.CreatedAt,
			})
			client.Send <- historyJSON
		}
	}

	notification, _ := json.Marshal(map[string]any{
		"type":   "user_joined_room",
		"room":   RoomID,
		"userId": client.ID,
	})

	room.Broadcast(notification, client)

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
	defer h.Mu.Unlock()

	room, exists := h.Rooms[RoomID]

	if !exists {
		return
	}

	room.RemoveClient(client)

	if room.IsEmpty() {
		delete(h.Rooms, RoomID)
	}
	client.Mu.Lock()
	delete(client.Rooms, RoomID)
	client.Mu.Unlock()

	notification, _ := json.Marshal(map[string]any{
		"type":   "user_left_room",
		"room":   RoomID,
		"userID": client.ID,
	})

	room.Broadcast(notification, nil)
}
