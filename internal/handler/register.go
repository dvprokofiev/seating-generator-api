package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dvprokofiev/seating-generator-api/internal/service"
)

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail), errors.Is(err, service.ErrPasswordTooShort):
			sendError(w, http.StatusBadRequest, "Incorrect email or password too short")

		case errors.Is(err, service.ErrUserAlreadyExists):
			sendError(w, http.StatusConflict, "User with such email already exists")

		default:
			sendError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}
