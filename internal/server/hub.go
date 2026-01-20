package server

import (
	"github.com/mildlybrutal/websocketGo/internal/common"
)

type MyServerHub struct {
	*common.Hub
}

type MyServerRoom struct {
	*common.Room
}

func (r *MyServerRoom) Broadcast(message []byte, exclude *common.Client) {
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

func (r *MyServerRoom) IsEmpty() bool {
	r.Mu.RLock()
	defer r.Mu.Unlock()

	return len(r.Clients) == 0
}
