package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/mildlybrutal/websocketGo/internal/middleware"
	"github.com/mildlybrutal/websocketGo/internal/repository"
	"github.com/mildlybrutal/websocketGo/internal/server"
	"github.com/mildlybrutal/websocketGo/internal/storage"
)

func main() {

	cfg, _ := common.LoadConfig(".")

	rdb, err := storage.InitRedis(&cfg.Redis)

	if err != nil {
		log.Fatalf("Redis init failed: %v", err)
	}

	db, err := storage.NewConnection(&cfg.Database)

	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	chatRepo := repository.NewChatRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	userRepo := repository.NewUserRepository(db)

	server.MainHub.ChatRepo = chatRepo
	server.MainHub.RedisClient = rdb

	go server.MainHub.Run()

	http.HandleFunc("/sign-up", server.SignUpHandler(userRepo))
	http.HandleFunc("/login", server.LoginHandler(userRepo, cfg))
	//protected
	http.HandleFunc("/ws", server.HandleConnections)

	http.HandleFunc("/api/profile", middleware.AuthMiddleware(
		server.GetUserProfileHandler(userRepo),
	))

	http.HandleFunc("/api/room/history", middleware.AuthMiddleware(
		server.GetRoomHistoryHandler(chatRepo),
	))

	http.HandleFunc("/api/room/create", middleware.AuthMiddleware(
		server.CreateRoomHandler(roomRepo),
	))

	http.HandleFunc("/api/auth/refresh", middleware.AuthMiddleware(
		server.RefreshTokenHandler(),
	))

	fmt.Println("Websocket server started at port 8080")

	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}
