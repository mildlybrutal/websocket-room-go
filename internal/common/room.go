package common

import (
	"sync"
)

type Room struct {
	ID      string
	Name    string
	Clients map[*Client]bool
	Mu      sync.RWMutex
}

func (r *Room) Broadcast(message []byte, exclude *Client) {
	//broadcasting logic
	r.Mu.RLock() // RLock because shared access (for reading only)
	defer r.Mu.RUnlock()

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

func (r *Room) AddClient(client *Client) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	if r.Clients == nil {
		r.Clients = make(map[*Client]bool)
	}

	r.Clients[client] = true
}

func (r *Room) RemoveClient(client *Client) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	delete(r.Clients, client)
}

func (r *Room) GetMemberIDs() []string {
	r.Mu.RLock()
	defer r.Mu.RUnlock()

	members := make([]string, 0, len(r.Clients))

	for client := range r.Clients {
		members = append(members, client.ID)
	}

	return members
}

func (r *Room) IsEmpty() bool {
	r.Mu.RLock()
	defer r.Mu.RUnlock()

	return len(r.Clients) == 0
}
