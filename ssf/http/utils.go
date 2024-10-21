package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

func GenerateStreamID() string {
	return fmt.Sprintf("stream-%d", time.Now().UnixNano())
}

func GenerateSecureEventToken(eventPayload map[string]interface{}, secret string) (string, error) {
	// Create a new token using the HS256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(eventPayload))

	// Sign the token using the secret
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func parseJWT(r *http.Request, secret string) (jwt.MapClaims, error) {
	// Read the request body to get the JWT token
	tokenString, err := io.ReadAll(r.Body) // `io.ReadAll` if you're using Go 1.16+
	if err != nil {
		return nil, err
	}

	// Parse the JWT token
	token, err := jwt.Parse(string(tokenString), func(token *jwt.Token) (interface{}, error) {
		// Ensure that the signing method is HMAC (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract and return the claims (payload)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}
