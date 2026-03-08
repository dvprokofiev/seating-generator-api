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

func TestAuthHandler(t *testing.T) {
	mockAuthSvc := service.NewMockAuthService(t)
	mockRegSvc := service.NewMockRegistrationService(t)
	h := NewAuthHandler(mockAuthSvc, mockRegSvc, nil)

	t.Run("success_login_200", func(t *testing.T) {
		testEmail := "valid@test.ru"
		testPass := "password123"
		fakeToken := "fake-jwt-token"

		mockAuthSvc.On("Login", mock.Anything, testEmail, testPass).Return(fakeToken, nil).Once()

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": testPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]string
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, fakeToken, resp["token"])
	})

	t.Run("wrong_password_401", func(t *testing.T) {
		testEmail := "valid@test.ru"
		testPass := "wrong-password"

		mockAuthSvc.On("Login", mock.Anything, testEmail, testPass).Return("", service.ErrInvalidCredentials).Once()

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": testPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("db_error_500", func(t *testing.T) {
		testEmail := "test@test.ru"
		testPass := "password"

		mockAuthSvc.On("Login", mock.Anything, testEmail, testPass).Return("", errors.New("db connection lost")).Once()

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": testPass,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("success_register_201", func(t *testing.T) {
		testEmail := "new@test.ru"
		testPass := "securepassword"

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
}
