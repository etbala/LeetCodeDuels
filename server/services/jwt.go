package services

import (
	"context"
	"errors"
	"fmt"
	"leetcodeduels/config"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID int64 `json:"userid"`
	jwt.RegisteredClaims
}

type contextKey string

const UserContextKey contextKey = "user"

// GenerateJWT generates a JWT for the given user ID and username.
// It returns the signed token as a string.
func GenerateJWT(userID int64) (string, error) {
	cfg := config.GetConfig()
	secretKey := cfg.JWT_SECRET

	// Create the JWT claims, including the user ID and username.
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 5 * time.Hour)), // 5 Day Lifetime
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates the given JWT string and returns the claims if valid.
// It returns an error if the token is invalid or expired.
func ValidateJWT(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()
	secretKey := cfg.JWT_SECRET

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func extractTokenString(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Missing Authorization Header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("Invalid Authorization Header Format")
	}

	return parts[1], nil
}

func GetClaimsFromRequest(r *http.Request) (*Claims, error) {
	tokenString, err := extractTokenString(r)
	if err != nil {
		return nil, err
	}
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// Middleware validates the JWT and attaches user information to the request context.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString, err := extractTokenString(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Attach the user information to the request context.
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
