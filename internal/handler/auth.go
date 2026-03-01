package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/dvprokofiev/seating-generator-api/internal/service"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{authService: s}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			sendError(w, http.StatusUnauthorized, "Неверная почта или пароль")
		default:
			log.Printf("Login error: %v", err)
			sendError(w, http.StatusInternalServerError, "Внутренняя ошибка сервера")
		}
		return
	}

	sendJSON(w, http.StatusOK, loginResponse{Token: token})
}
