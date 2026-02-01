package server

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateToken(tokenStr string, secret string) (uint, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}

	userID := uint(claims["sub"].(float64))

	return userID, nil
}
