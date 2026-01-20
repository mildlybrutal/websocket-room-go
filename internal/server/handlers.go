package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mildlybrutal/websocketGo/internal/common"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var MainHub = common.NewHub()

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	clientID := r.URL.Query().Get("id")

	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	baseClient := &common.Client{
		ID:    clientID,
		Conn:  conn,
		Send:  make(chan []byte, 256),
		Hub:   MainHub,
		Rooms: make(map[string]bool),
	}

	serverClient := &MyServerClient{
		Client: baseClient,
	}

	serverClient.Hub.Register <- baseClient

	go serverClient.WritePump()
	go serverClient.ReadPump()
}
