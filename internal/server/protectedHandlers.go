package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/mildlybrutal/websocketGo/internal/middleware"
	"github.com/mildlybrutal/websocketGo/internal/repository"
)

// GetUserProfileHandler returns the authenticated user's profile
func GetUserProfileHandler(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID from context
		userID, ok := middleware.RequireAuth(r, w)
		if !ok {
			return
		}

		// Fetch user from database
		user, err := userRepo.GetByID(userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Return user profile
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		})
	}
}

// CreateRoomHandler creates a new chat room (protected)
func CreateRoomHandler(roomRepo *repository.RoomRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.RequireAuth(r, w)
		if !ok {
			return
		}

		var req struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "Room name is required", http.StatusBadRequest)
			return
		}

		// Create room with authenticated user as owner
		if err := roomRepo.CreateRoom(req.Name, userID); err != nil {
			log.Printf("Failed to create room: %v", err)
			http.Error(w, "Failed to create room", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Room created successfully",
			"name":    req.Name,
		})
	}
}

// GetRoomHistoryHandler returns chat history for a room (protected)
func GetRoomHistoryHandler(chatRepo *repository.ChatRepository, roomRepo *repository.RoomRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.RequireAuth(r, w)
		if !ok {
			return
		}

		// Get room ID from URL query parameter
		roomIDStr := r.URL.Query().Get("room_id")
		if roomIDStr == "" {
			http.Error(w, "room_id parameter required", http.StatusBadRequest)
			return
		}

		roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid room_id", http.StatusBadRequest)
			return
		}

		// TODO: Check if user has access to this room
		hasAccess, err := roomRepo.CheckUserRoomAccess(userID, uint(roomID))

		if err != nil {
			log.Printf("Error checking room access: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !hasAccess {
			http.Error(w, "You don't have access to this room", http.StatusForbidden)
			return
		}

		// Get message limit from query (default 50, max 100)
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				if l > 0 && l <= 100 {
					limit = l
				}
			}
		}

		// Fetch history
		messages, err := chatRepo.GetRoomHistory(uint(roomID), limit)
		if err != nil {
			log.Printf("Failed to get room history: %v", err)
			http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"room_id":  roomID,
			"messages": messages, // ✓ FIXED: Now correct type
			"count":    len(messages),
		})
	}
}

// RefreshTokenHandler generates a new JWT token (requires valid token)
func RefreshTokenHandler(cfg *common.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r)
		if !ok || userID == 0 {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		username, _ := middleware.GetUsernameFromContext(r)

		token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
			"sub":      userID,
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
			"iat":      time.Now().Unix(),
		})

		tokenString, err := token.SignedString([]byte(cfg.Security.JWTSecret))

		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]any{
			"token":      tokenString,
			"expires_in": 86400, // 24 hours
			"user_id":    userID,
			"username":   username,
		})
	}
}
