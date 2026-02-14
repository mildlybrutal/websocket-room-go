package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/mildlybrutal/websocketGo/internal/middleware"
	"github.com/mildlybrutal/websocketGo/internal/server/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var MainHub = common.NewHub()

type UserRepo interface {
	GetByID(uint) (*models.User, error)
}

func HandleConnections(userRepo UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, _ := common.LoadConfig(".")

		var userID uint
		var username string = "Anonymous"

		token := r.URL.Query().Get("token")
		if token != "" {
			id, err := middleware.ValidateToken(token, cfg.Security.JWTSecret)
			if err == nil {
				userID = id
				// Fetch username
				user, err := userRepo.GetByID(userID)
				if err == nil {
					username = user.Username
				}
			} else {
				log.Printf("Invalid token: %v", err)
			}
		}

		if cfg.Security.RequireAuth && userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
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
			ID:       clientID,
			UserID:   userID,
			Username: username,
			Conn:     conn,
			Send:     make(chan []byte, 256),
			Hub:      MainHub,
			Rooms:    make(map[string]bool),
		}

		serverClient := NewServerClient(baseClient)

		serverClient.Hub.Register <- baseClient

		go serverClient.WritePump()
		go serverClient.ReadPump()
	}
}
