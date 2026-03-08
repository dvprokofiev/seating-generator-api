package handler

import (
	"github.com/dvprokofiev/seating-generator-api/internal/service"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService          service.AuthService
	registrationService  service.RegistrationService
	emailVerifierService service.EmailVerifier
	validator            *validator.Validate
}

func NewAuthHandler(s service.AuthService, r service.RegistrationService, e service.EmailVerifier) *AuthHandler {
	return &AuthHandler{
		authService:          s,
		registrationService:  r,
		emailVerifierService: e,
		validator:            validator.New(),
	}
}
