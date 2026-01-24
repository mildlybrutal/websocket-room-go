package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/mildlybrutal/websocketGo/internal/repository"
	"github.com/mildlybrutal/websocketGo/internal/server"
	"github.com/mildlybrutal/websocketGo/internal/storage"
)

func main() {

	cfg, _ := common.LoadConfig(".")

	db, err := storage.NewConnection(&cfg.Database)

	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	chatRepo := repository.NewChatRepository(db)

	server.MainHub.ChatRepo = chatRepo

	go server.MainHub.Run()

	http.HandleFunc("/ws", server.HandleConnections)

	fmt.Println("Websocket server started at port 8080")

	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}
