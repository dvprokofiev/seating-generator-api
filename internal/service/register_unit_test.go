package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthService_Register_Unit(t *testing.T) {
	mockRepo := new(repository.MockUserRepository)
	svc := NewAuthService(mockRepo, "secret")

	t.Run("successful_registration", func(t *testing.T) {
		email := "newuser@test.ru"
		password := "password123"

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil).Once()

		err := svc.Register(context.Background(), email, password)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		email := "existing@test.ru"
		password := "password123"

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(repository.ErrDuplicateEmail).Once()

		err := svc.Register(context.Background(), email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserAlreadyExists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid_email", func(t *testing.T) {
		localMock := new(repository.MockUserRepository)
		localSvc := NewAuthService(localMock, "secret")

		email := "not_an_email"
		password := "password123"

		err := localSvc.Register(context.Background(), email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEmail)
		localMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("short_password", func(t *testing.T) {
		localMock := new(repository.MockUserRepository)
		localSvc := NewAuthService(localMock, "secret")

		email := "test@test.ru"
		shortPass := "12345"

		err := localSvc.Register(context.Background(), email, shortPass)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPasswordTooShort)
		localMock.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("db_error", func(t *testing.T) {
		email := "test@test.ru"
		password := "password123"

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(errors.New("database connection lost")).Once()

		err := svc.Register(context.Background(), email, password)

		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrUserAlreadyExists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("email_normalized_to_lowercase", func(t *testing.T) {
		email := "UPPERCASE@TEST.RU"
		password := "password123"
		var capturedUser *models.User

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
			capturedUser = user
			return true
		})).Return(nil).Once()

		err := svc.Register(context.Background(), email, password)

		assert.NoError(t, err)
		assert.Equal(t, "uppercase@test.ru", capturedUser.Email)
		mockRepo.AssertExpectations(t)
	})
}
