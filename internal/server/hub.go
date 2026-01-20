package server

import (
	"encoding/json"
	"log"
	"time"

	"github.com/mildlybrutal/websocketGo/internal/common"
)

type MyServerHub struct {
	*common.Hub
}

type MyServerRoom struct {
	*common.Room
}

func NewHub() *common.Hub {
	return &common.Hub{
		Clients:    make(map[*common.Client]bool),
		Rooms:      make(map[string]*common.Room),
		Broadcast:  make(chan common.BroadcastMessage),
		Register:   make(chan *common.Client),
		Unregister: make(chan *common.Client),
	}
}

func (h *MyServerHub) Run() {
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
			if _, ok := h.Clients[client]; ok {
				for roomID := range h.Rooms {
					if room, exists := h.Rooms[roomID]; exists {
						room.RemoveClient(client)
					}
				}
				delete(h.Clients, client.ID) //remove from list
				close(client.Send)           //shuts down client's individual channel
				log.Printf("Client %s unregistered", client.ID)

				h.Mu.Unlock() // Unlock before broadcasting
			} else {
				h.Mu.Unlock()
			}
		case message := <-h.Broadcast:
			// A message was received from one client that needs to go to everyone.
			h.broadcastMessage(message)
		case <-ticker.C:
			h.cleanup()
		}
	}
}

func (r *MyServerRoom) Broadcast(message []byte, exclude *common.Client) {
	//broadcasting logic
	r.Mu.RLock() // RLock because shared access (for reading only)
	defer r.Mu.RLock()

	for client := range r.Clients {
		//exclude -> send a message to "everyone but me."
		if client != exclude {
			select {
			case client.Send <- message:
			default:
				//Buffer full,skip
			}
		}
	}
}

func (h *MyServerHub) JoinRoom(RoomID string, client *common.Client) error {
	h.Mu.Lock()
	room, exists := h.Rooms[RoomID]
	if !exists {
		room = &common.Room{
			ID:      RoomID,
			Clients: make(map[*common.Client]bool),
		}

		h.Rooms[RoomID] = room
	}

	h.Mu.Unlock()

	room.AddClient(client)
	client.Broadcast[room]

	notification, _ := json.Marshal(map[string]interface{}{
		"type":   "user_joined_room",
		"room":   RoomID,
		"userId": client.ID,
	})

	room.Broadcast(notification, client)
	roomInfo, _ := json.Marshal(map[string]interface{}{
		"type":    "room_joined",
		"room":    RoomID,
		"members": room.GetMemberIDs(),
	})

	client.Send <- roomInfo

	return nil
}

func (h *MyServerHub) LeaveRoom(client *common.Client, RoomID string) {
	h.Mu.RLock()

	room, exists := h.Rooms[RoomID]
	h.Mu.RUnlock()

	if !exists {
		return
	}

	room.RemoveClient(client)

	delete(client.Rooms, RoomID)

	notification, _ := json.Marshal(map[string]interface{}{
		"type":   "user_left_room",
		"room":   RoomID,
		"userID": client.ID,
	})

	room.Broadcast(notification, nil)

	if room.IsEmpty() {
		h.Mu.Lock()
		delete(h.Rooms, RoomID)
		h.Mu.Unlock()
	}
}

func (r *MyServerRoom) AddClient(client *common.Client) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	r.Clients[client] = true
}

func (r *MyServerRoom) RemoveClient(client *common.Client) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	delete(r.Clients, client)
}

func (r *MyServerRoom) GetMemberIDs() []string {
	r.Mu.RLock()
	defer r.Mu.RUnlock()

	members := make([]string, 0, len(r.Clients))

	for client := range r.Clients {
		members = append(members, client.ID)
	}

	return members
}

func (r *MyServerRoom) IsEmpty() {
	r.Mu.RLock()
	defer r.Mu.Unlock()

	return len(r.Clients) == 0
}
