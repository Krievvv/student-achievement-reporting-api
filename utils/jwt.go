package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Return: (AccessToken, RefreshToken, Error)
func GenerateTokens(userID, username, role, roleID string) (string, string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	// 1. Access Token Claims
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"role_id":  roleID,
		"exp":      time.Now().Add(time.Hour * 10).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token Claims
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	rt, err := refreshToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	return t, rt, nil
}