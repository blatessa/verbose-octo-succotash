package auth

import (
	"context"
	"errors"
	"fmt"

	authdb "github.com/blatessa/verbose-octo-succotash/internal/auth/db"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("auth: invalid credentials")

// User is the auth package's domain type for a registered user.
// It never exposes the password hash.
type User struct {
	ID    string
	Email string
}

type Service struct {
	q   *authdb.Queries
	cfg Config
}

func NewService(q *authdb.Queries, cfg Config) *Service {
	return &Service{q: q, cfg: cfg}
}

// CreateUser registers a new user and returns the created User plus a signed JWT.
func (s *Service) CreateUser(ctx context.Context, email, password string) (User, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, "", fmt.Errorf("auth: hash password: %w", err)
	}

	row, err := s.q.CreateUser(ctx, authdb.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return User{}, "", fmt.Errorf("auth: create user: %w", err)
	}

	u := User{ID: row.ID.String(), Email: row.Email}
	token, err := GenerateToken(u.ID, u.Email, s.cfg)
	if err != nil {
		return User{}, "", err
	}
	return u, token, nil
}

// Login validates credentials and returns the User plus a signed JWT.
func (s *Service) Login(ctx context.Context, email, password string) (User, string, error) {
	row, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return User{}, "", ErrInvalidCredentials
	}

	u := User{ID: row.ID.String(), Email: row.Email}
	token, err := GenerateToken(u.ID, u.Email, s.cfg)
	if err != nil {
		return User{}, "", err
	}
	return u, token, nil
}
