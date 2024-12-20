package auth

import (
	"fmt"
	"leetcodeduels/pkg/config"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   int64  `json:"sub"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT for the given user ID and username.
// It returns the signed token as a string.
func GenerateJWT(userID int64, username string) (string, error) {
	cfg, _ := config.LoadConfig()
	secretKey := cfg.JWT_SECRET

	// Create the JWT claims, including the user ID and username.
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 Hour Lifetime
		},
	}

	// Create the token with the specified claims and signing method.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key.
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates the given JWT string and returns the claims if valid.
// It returns an error if the token is invalid or expired.
func ValidateJWT(tokenString string) (*Claims, error) {
	cfg, _ := config.LoadConfig()
	secretKey := cfg.JWT_SECRET

	// Initialize a new instance of Claims.
	claims := &Claims{}

	// Parse the token, validating the signature and claims.
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC and using SHA256.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if the token is valid.
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
