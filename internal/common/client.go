package common

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	UserID   uint
	Username string
	Conn     *websocket.Conn
	Send   chan []byte
	Hub    *Hub
	Rooms  map[string]bool
	Mu     sync.RWMutex
}
