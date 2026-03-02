package service

import (
	"context"
	"errors"

	"github.com/dvprokofiev/seating-generator-api/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrInternal           = errors.New("Internal server error")
	ErrPasswordTooShort   = errors.New("Password is less then 8 symbols length")
	ErrInvalidEmail       = errors.New("Invalid email")
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
}

type authService struct {
	repo      repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(repo repository.UserRepository, secret string) AuthService {
	return &authService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}
