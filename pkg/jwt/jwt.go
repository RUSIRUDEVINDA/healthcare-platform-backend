package jwt

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT token payload
// This is embedded in every access token
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Helper wraps JWT operations
// Any service that needs to verify tokens imports this pkg
type Helper struct {
	accessSecret   []byte
	refreshSecret  []byte
	AccessTTLMins  int
	RefreshTTLDays int
}

func New(accessSecret, refreshSecret string, accessTTLMins, refreshTTLDays int) *Helper {
	return &Helper{
		accessSecret:   []byte(accessSecret),
		refreshSecret:  []byte(refreshSecret),
		AccessTTLMins:  accessTTLMins,
		RefreshTTLDays: refreshTTLDays,
	}
}

// GenerateAccessToken creates a signed JWT access token
// Returns: tokenString, expiresInSeconds, error
func (h *Helper) GenerateAccessToken(userID, email, role string) (string, int, error) {
	expiresAt := time.Now().UTC().Add(time.Duration(h.AccessTTLMins) * time.Minute)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "healthcare-platform",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(h.accessSecret)
	if err != nil {
		return "", 0, fmt.Errorf("jwt.GenerateAccessToken: %w", err)
	}

	return signed, h.AccessTTLMins * 60, nil
}

// ParseAccessToken validates and parses an access token
// Other services can use this to validate tokens without calling auth-service
func (h *Helper) ParseAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.accessSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token has expired")
		}
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// HashToken creates a SHA-256 hash of a token for safe database storage
// We never store plaintext refresh tokens — only their hashes
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
