package server

import (
	"fmt"
	"log"
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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	//upgrading websockets
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("Failed to upgrade: %v", err)
		return
	}

	defer conn.Close()

	welcome := common.Message{
		Type:    "welcome",
		Content: "Connected to WebSocket server",
	}

	err = conn.WriteJSON(welcome)

	if err := conn.WriteJSON(welcome); err != nil {
		log.Printf("Error sending welcome: %v", err)
		return
	}

	for {
		var msg common.Message
		err := conn.ReadJSON(&msg)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("Websocket error: %v", err)
			}
			break
		}

		log.Printf("Recieved: %+v", msg)

		response := common.Message{
			Type:    "echo",
			Content: msg.Content,
		}

		if err := conn.WriteJSON(response); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
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
