package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mildlybrutal/websocketGo/internal/common"
)

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UsernameKey contextKey = "username"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Auth header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format. Use: Bearer <token>", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		cfg, err := common.LoadConfig(".")

		if err != nil {
			log.Printf("Failed to load config in auth middleware: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		userID, username, err := ValidateTokenWithClaims(tokenString, cfg.Security.JWTSecret)

		if err != nil {
			log.Printf("Token validation failed: %v", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, userID)
		ctx = context.WithValue(ctx, UsernameKey, username)

		next.ServeHTTP(w, r.WithContext(ctx))

	}
}

func ValidateToken(tokenStr string, secret string) (uint, error) {
	userID, _, err := ValidateTokenWithClaims(tokenStr, secret)
	return userID, err
}

func ValidateTokenWithClaims(tokenStr string, secret string) (uint, string, error) {
	//parse token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return 0, "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", fmt.Errorf("invalid claims format")
	}

	// Extract userID
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, "", fmt.Errorf("invalid user ID in token")
	}
	userID := uint(userIDFloat)

	// Extract username (optional)
	username, _ := claims["username"].(string)

	return userID, username, nil
}

func GetUserIDFromContext(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uint)
	return userID, ok
}

func GetUsernameFromContext(r *http.Request) (string, bool) {
	username, ok := r.Context().Value(UsernameKey).(string)
	return username, ok
}

func RequireAuth(r *http.Request, w http.ResponseWriter) (uint, bool) {
	userID, ok := GetUserIDFromContext(r)
	if !ok || userID == 0 {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}
