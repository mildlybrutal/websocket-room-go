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
	cfg, _ := common.LoadConfig(".")

	var userID uint

	if cfg.Security.RequireAuth {
		token := r.URL.Query().Get("token")
		id, err := ValidateToken(token, cfg.Security.JWTSecret)

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userID = id
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	clientID := r.URL.Query().Get("id")

	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	baseClient := &common.Client{
		ID:     clientID,
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    MainHub,
		Rooms:  make(map[string]bool),
	}

	serverClient := &MyServerClient{
		Client: baseClient,
	}

	serverClient.Hub.Register <- baseClient

	go serverClient.WritePump()
	go serverClient.ReadPump()
}
