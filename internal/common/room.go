package common

import "sync"

type Room struct {
	ID      string
	Name    string
	Clients map[*Client]bool
	Mu      sync.RWMutex
}
