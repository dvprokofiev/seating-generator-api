package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/dvprokofiev/seating-generator-api/internal/service"
	"github.com/go-playground/validator/v10"
)

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginResponse struct {
	Token string `json:"token"`
}

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

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		sendError(w, http.StatusBadRequest, "Validation failed "+err.Error())
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			sendError(w, http.StatusUnauthorized, "Incorrect e-mail or password")
		default:
			log.Printf("Login error: %v", err)
			sendError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	sendJSON(w, http.StatusOK, loginResponse{Token: token})
}
