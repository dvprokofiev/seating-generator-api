package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/dvprokofiev/seating-generator-api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthHandler_WithMockRepo(t *testing.T) {
	mockRepo := repository.NewMockUserRepository(t)
	authSvc := service.NewAuthService(mockRepo, "super-secret")
	h := NewAuthHandler(authSvc)

	t.Run("success_login_200", func(t *testing.T) {
		testEmail := "valid@test.ru"
		testPass := "password123"

		hash, _ := bcrypt.GenerateFromPassword([]byte(testPass), bcrypt.DefaultCost)
		fakeUser := &models.User{
			ID:           uuid.New(),
			Email:        testEmail,
			PasswordHash: string(hash),
		}

		mockRepo.On("GetByEmail", mock.Anything, testEmail).Return(fakeUser, nil).Once()

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
		assert.NotEmpty(t, resp["token"], "Token should not be empty on success")
	})

	t.Run("wrong_password_401", func(t *testing.T) {
		testEmail := "valid@test.ru"
		hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
		fakeUser := &models.User{
			Email:        testEmail,
			PasswordHash: string(hash),
		}

		mockRepo.On("GetByEmail", mock.Anything, testEmail).Return(fakeUser, nil).Once()

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": "WRONG-password",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("db_error_500", func(t *testing.T) {
		testEmail := "test@test.ru"
		mockRepo.On("GetByEmail", mock.Anything, testEmail).
			Return(nil, errors.New("database connection lost")).Once()

		body, _ := json.Marshal(map[string]string{
			"email":    testEmail,
			"password": "password",
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("user_not_found_returns_401", func(t *testing.T) {
		testEmail := "notfound@mail.ru"
		mockRepo.On("GetByEmail", mock.Anything, testEmail).Return(nil, nil).Once()

		body, _ := json.Marshal(map[string]string{"email": testEmail, "password": "password"})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
