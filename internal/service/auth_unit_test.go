package service

import (
	"context"
	"testing"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Login_UUID(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	svc := NewAuthService(mockRepo, "secret")

	t.Run("successful_login", func(t *testing.T) {
		email := "test@test.ru"
		pass := "password"
		userID := uuid.New()

		hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

		existingUser := &models.User{
			ID:           userID,
			Email:        email,
			PasswordHash: string(hash),
		}

		mockRepo.On("GetByEmail", mock.Anything, email).Return(existingUser, nil).Once()

		token, err := svc.Login(context.Background(), email, pass)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockRepo.AssertExpectations(t)
	})
	t.Run("wrong_email", func(t *testing.T) {
		email := "non-existent@test.ru"
		mockRepo.On("GetByEmail", mock.Anything, email).
			Return(nil, assert.AnError).Once()

		token, err := svc.Login(context.Background(), email, "any-password")

		assert.Error(t, err)
		assert.Empty(t, token)
		mockRepo.AssertExpectations(t)
	})
	t.Run("wrong_password", func(t *testing.T) {
		email := "user@test.ru"
		realPassword := "correct-password"
		wrongPassword := "incorrect-password"
		hash, _ := bcrypt.GenerateFromPassword([]byte(realPassword), bcrypt.DefaultCost)

		existingUser := &models.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hash),
		}
		mockRepo.On("GetByEmail", mock.Anything, email).Return(existingUser, nil).Once()

		token, err := svc.Login(context.Background(), email, wrongPassword)
		assert.Error(t, err)
		assert.Empty(t, token)

		mockRepo.AssertExpectations(t)
	})
	t.Run("short_password", func(t *testing.T) {
		localMock := new(repository.MockUserRepository)
		localSvc := NewAuthService(localMock, "secret")

		email := "test@test.ru"
		shortPass := "12345"

		token, err := localSvc.Login(context.Background(), email, shortPass)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, ErrPasswordTooShort)
		localMock.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything)
	})
	t.Run("incorrect_email", func(t *testing.T) {
		localMock := new(repository.MockUserRepository)
		localSvc := NewAuthService(localMock, "secret")

		email := "not_an_email"
		shortPass := "12345678"

		token, err := localSvc.Login(context.Background(), email, shortPass)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, ErrInvalidEmail)
		localMock.AssertNotCalled(t, "GetByEmail", mock.Anything, mock.Anything)
	})
}
