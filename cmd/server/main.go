package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/mildlybrutal/websocketGo/internal/middleware"
	"github.com/mildlybrutal/websocketGo/internal/repository"
	"github.com/mildlybrutal/websocketGo/internal/server"
	"github.com/mildlybrutal/websocketGo/internal/server/models"
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

	// Seed Data
	systemUser, err := userRepo.GetOrCreateUser("system", "system@localhost")
	if err != nil {
		log.Printf("Failed to create system user: %v", err)
	} else {
		rooms := []string{"General", "Random", "Tech"}
		for _, roomName := range rooms {
			var room models.Room
			if err := db.Where("name = ?", roomName).First(&room).Error; err != nil {
				if err := roomRepo.CreateRoom(roomName, systemUser.ID); err != nil {
					log.Printf("Failed to seed room %s: %v", roomName, err)
				} else {
					log.Printf("Seeded room: %s", roomName)
				}
			}
		}
	}

	server.MainHub.ChatRepo = chatRepo
	server.MainHub.RedisClient = rdb

	go server.MainHub.Run()

	http.HandleFunc("/sign-up", server.SignUpHandler(userRepo))
	http.HandleFunc("/login", server.LoginHandler(userRepo, cfg))
	//protected
	http.HandleFunc("/ws", server.HandleConnections(userRepo))

	http.HandleFunc("/api/profile", middleware.AuthMiddleware(
		server.GetUserProfileHandler(userRepo),
	))

	http.HandleFunc("/api/room/history", middleware.AuthMiddleware(
		server.GetRoomHistoryHandler(chatRepo, roomRepo),
	))

	http.HandleFunc("/api/room/create", middleware.AuthMiddleware(
		server.CreateRoomHandler(roomRepo),
	))

	http.HandleFunc("/api/rooms", middleware.AuthMiddleware(
		server.GetRoomsHandler(roomRepo),
	))

	http.HandleFunc("/api/room/join", middleware.AuthMiddleware(
		server.JoinRoomHandler(roomRepo),
	))

	http.HandleFunc("/api/auth/refresh", middleware.AuthMiddleware(
		server.RefreshTokenHandler(cfg),
	))

	handler := middleware.CORSMiddleware(cfg)(http.DefaultServeMux)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("WebSocket server started at http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-stop
	log.Println("Shutting down server")

	server.MainHub.Shutdown()
	log.Println("hub shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")

}
