package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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
func CreateRoomHandler(roomRepo interface {
	CreateRoom(name string, ownerID uint) error
}) http.HandlerFunc {
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
func GetRoomHistoryHandler(chatRepo interface {
	GetRoomHistory(roomID uint, limit int) ([]any, error)
}) http.HandlerFunc {
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
		// hasAccess := checkRoomAccess(userID, uint(roomID))

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
			"messages": messages,
			"count":    len(messages),
		})
	}
}

// RefreshTokenHandler generates a new JWT token (requires valid token)
func RefreshTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserIDFromContext(r)
		if !ok || userID == 0 {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		username, _ := middleware.GetUsernameFromContext(r)

		// Generate new token
		// ... (use the token generation logic from LoginHandler)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"message":  "Token refreshed",
			"user_id":  userID,
			"username": username,
		})
	}
}
