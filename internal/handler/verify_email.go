package handler

import (
	"errors"
	"net/http"

	"github.com/dvprokofiev/seating-generator-api/internal/service"
	"github.com/google/uuid"
)

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		sendError(w, http.StatusBadRequest, "Verification token is required")
		return
	}
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid token format")
		return
	}

	ctx := r.Context()
	err = h.emailVerifierService.Verify(ctx, token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTokenNotFound):
			sendError(w, http.StatusUnauthorized, "Token was not found")
		case errors.Is(err, service.ErrTokenExpired):
			sendError(w, http.StatusUnauthorized, "Token has expired")
		default:
			sendError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "Email verified"})
}
