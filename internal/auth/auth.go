package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const claimsKey contextKey = "auth_claims"

// ClaimsFromContext retrieves the JWT Claims stored by Middleware.
func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(claimsKey).(Claims)
	return c, ok
}

var (
	ErrTokenExpired = errors.New("auth: token expired")
	ErrTokenInvalid = errors.New("auth: token invalid")
)

// Claims holds the data embedded in a JWT.
type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Expiry int64  `json:"exp"` // Unix timestamp
}

// Config holds signing parameters.
type Config struct {
	Secret     []byte
	Expiration time.Duration
}

var b64 = base64.RawURLEncoding

var jwtHeader = b64.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

// GenerateToken signs a HS256 JWT for the given user.
func GenerateToken(userID, email string, cfg Config) (string, error) {
	if len(cfg.Secret) == 0 {
		return "", errors.New("auth: secret must not be empty")
	}
	claims := Claims{
		UserID: userID,
		Email:  email,
		Expiry: time.Now().Add(cfg.Expiration).Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	body := jwtHeader + "." + b64.EncodeToString(payload)
	sig := sign(cfg.Secret, body)
	return body + "." + sig, nil
}

// ValidateToken parses and verifies a HS256 JWT, returning its claims.
func ValidateToken(token string, cfg Config) (Claims, error) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return Claims{}, ErrTokenInvalid
	}
	body := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(sign(cfg.Secret, body)), []byte(parts[2])) {
		return Claims{}, ErrTokenInvalid
	}
	raw, err := b64.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrTokenInvalid
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return Claims{}, ErrTokenInvalid
	}
	if time.Now().Unix() > c.Expiry {
		return Claims{}, ErrTokenExpired
	}
	return c, nil
}

// Middleware validates the Bearer JWT on every request before calling next.
func Middleware(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := ValidateToken(strings.TrimPrefix(header, "Bearer "), cfg)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GenerateSecret returns a cryptographically random 32-byte secret suitable
// for use as a signing key.
func GenerateSecret() ([]byte, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return b, err
}

func sign(secret []byte, data string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return b64.EncodeToString(mac.Sum(nil))
}
