package server

import (
	"log"
	"time"

	"github.com/mildlybrutal/websocketGo/internal/common"
)

type MyServerHub struct {
	*common.Hub
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

func (h *MyServerHub) broadcastMessage(message []byte, exclude *common.Client) {
	//broadcasting logic
	h.Mu.RLock() // RLock because shared access (for reading only)
	defer h.Mu.RLock()

	for client := range h.Clients {
		//exclude -> send a message to "everyone but me."
		if client != exclude {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.Clients, client)
			}
		}
	}
}
