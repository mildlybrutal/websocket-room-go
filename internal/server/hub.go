package server

import (
	"encoding/json"
	"log"

	"github.com/mildlybrutal/websocketGo/internal/common"
)

type MyServerHub struct {
	*common.Hub
}

func NewHub() *common.Hub {
	return &common.Hub{
		Clients:    make(map[*common.Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *common.Client),
		Unregister: make(chan *common.Client),
	}
}

func (h *MyServerHub) Run() {
	for {
		// select lets the hub listen to multiple channels at once.
		select {
		case client := <-h.Register:
			//New user connected
			h.Mu.Lock() // exclusive for the client
			h.Clients[client] = true
			h.Mu.Unlock()
			log.Printf("Client %s registered", client.ID)

			notification, _ := json.Marshal(map[string]string{
				"type": "user_joined",
				"id":   client.ID,
			})
			h.broadcastMessage(notification, client)
		case client := <-h.Unregister:
			//user disconnection
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client) //remove from list
				close(client.Send)        //shuts down client's individual channel
				log.Printf("Client %s unregistered", client.ID)

				notification, _ := json.Marshal(map[string]string{
					"type": "user_joined",
					"id":   client.ID,
				})
				h.Mu.Unlock() // Unlock before broadcasting
				h.broadcastMessage(notification, client)
			} else {
				h.Mu.Unlock()
			}
		case message := <-h.Broadcast:
			// A message was received from one client that needs to go to everyone.
			h.broadcastMessage(message, nil)
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
