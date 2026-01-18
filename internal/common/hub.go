package common

import "sync"

type Hub struct {
	Clients    map[*Client]bool
	Rooms      map[string]*Room
	Broadcast  chan BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
	Mu         sync.RWMutex
}
