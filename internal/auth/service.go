package auth

import (
	"context"
	"errors"
	"fmt"

	authdb "github.com/blatessa/verbose-octo-succotash/internal/auth/db"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("auth: invalid credentials")

type Service struct {
	q   *authdb.Queries
	cfg Config
}

func NewService(q *authdb.Queries, cfg Config) *Service {
	return &Service{q: q, cfg: cfg}
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// CreateUser registers a new user and returns a signed JWT.
func (s *Service) CreateUser(ctx context.Context, email, password string) (AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth: hash password: %w", err)
	}

	user, err := s.q.CreateUser(ctx, authdb.CreateUserParams{
		Email:        email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return AuthResponse{}, fmt.Errorf("auth: create user: %w", err)
	}

	id := fmt.Sprintf("%x-%x-%x-%x-%x",
		user.ID.Bytes[0:4], user.ID.Bytes[4:6],
		user.ID.Bytes[6:8], user.ID.Bytes[8:10],
		user.ID.Bytes[10:16],
	)
	token, err := GenerateToken(id, user.Email, s.cfg)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token: token,
		User:  UserResponse{ID: id, Email: user.Email},
	}, nil
}

// Login validates credentials and returns a signed JWT.
func (s *Service) Login(ctx context.Context, email, password string) (AuthResponse, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	id := fmt.Sprintf("%x-%x-%x-%x-%x",
		user.ID.Bytes[0:4], user.ID.Bytes[4:6],
		user.ID.Bytes[6:8], user.ID.Bytes[8:10],
		user.ID.Bytes[10:16],
	)
	token, err := GenerateToken(id, user.Email, s.cfg)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token: token,
		User:  UserResponse{ID: id, Email: user.Email},
	}, nil
}
