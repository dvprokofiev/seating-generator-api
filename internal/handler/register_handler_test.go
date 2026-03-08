package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dvprokofiev/seating-generator-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterHandler(t *testing.T) {
	mockAuthSvc := service.NewMockAuthService(t)
	mockRegSvc := service.NewMockRegistrationService(t)
	mockVerifier := service.NewMockEmailVerifier(t)
	h := NewAuthHandler(mockAuthSvc, mockRegSvc, mockVerifier)

	t.Run("success_registration_201", func(t *testing.T) {
		testEmail := "newuser@test.ru"
		testPass := "password123"

		mockRegSvc.On("Register", mock.Anything, testEmail, testPass).Return(nil).Once()

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": testPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("duplicate_email_409", func(t *testing.T) {
		testEmail := "existing@test.ru"
		testPass := "password123"

		mockRegSvc.On("Register", mock.Anything, testEmail, testPass).Return(service.ErrUserAlreadyExists).Once()

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": testPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("invalid_email_400", func(t *testing.T) {
		testEmail := "invalid-email"
		testPass := "password123"

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": testPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("short_password_400", func(t *testing.T) {
		testEmail := "test@test.ru"
		testPass := "12345"

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": testPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("db_error_500", func(t *testing.T) {
		testEmail := "test@test.ru"
		testPass := "password123"

		mockRegSvc.On("Register", mock.Anything, testEmail, testPass).Return(errors.New("database connection lost")).Once()

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": testPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("invalid_json_400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer([]byte("not json")))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing_email_400", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"password": "password123",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("missing_password_400", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email": "test@test.ru",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
