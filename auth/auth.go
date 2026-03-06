package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

// Claims represents the payload stored in a JWT token.
type Claims struct {
	UserID string
	Email  string
	Expiry time.Time
}

// TokenConfig holds the configuration for token signing.
type TokenConfig struct {
	Secret     string
	Expiration time.Duration
}

// GenerateToken creates a signed token for the given claims.
func GenerateToken(claims Claims, cfg TokenConfig) (string, error) {
	// TODO: implement JWT signing
	return "", errors.New("not implemented")
}

// ValidateToken parses and validates a token string, returning its claims.
func ValidateToken(token string, cfg TokenConfig) (Claims, error) {
	// TODO: implement JWT validation
	return Claims{}, errors.New("not implemented")
}

// Middleware extracts and validates the Bearer token from the Authorization
// header, rejecting requests that are missing or have an invalid token.
func Middleware(cfg TokenConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			_, err := ValidateToken(token, cfg)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
