package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dvprokofiev/seating-generator-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVerifyHandler(t *testing.T) {
	mockAuthSvc := service.NewMockAuthService(t)
	mockRegSvc := service.NewMockRegistrationService(t)
	mockEmailSvc := service.NewMockEmailVerifier(t)

	h := NewAuthHandler(mockAuthSvc, mockRegSvc, mockEmailSvc)
	t.Run("verify_success", func(t *testing.T) {
		token := uuid.New()

		mockEmailSvc.On("Verify", mock.Anything, token).Return(nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/auth/verify?token="+token.String(), nil)
		w := httptest.NewRecorder()

		h.VerifyEmail(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Email verified")
	})

	t.Run("no_such_token", func(t *testing.T) {
		token := uuid.New()

		mockEmailSvc.On("Verify", mock.Anything, token).Return(service.ErrTokenNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/auth/verify?token="+token.String(), nil)
		w := httptest.NewRecorder()

		h.VerifyEmail(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "not found")
	})

	t.Run("verify_expired_token", func(t *testing.T) {
		token := uuid.New()

		mockEmailSvc.On("Verify", mock.Anything, token).
			Return(service.ErrTokenExpired).Once()

		req := httptest.NewRequest(http.MethodGet, "/auth/verify?token="+token.String(), nil)
		w := httptest.NewRecorder()

		h.VerifyEmail(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "expired")
	})
}
