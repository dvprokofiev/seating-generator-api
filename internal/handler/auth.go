package handler

import (
	"github.com/dvprokofiev/seating-generator-api/internal/service"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService service.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: s,
		validator:   validator.New(),
	}
}
