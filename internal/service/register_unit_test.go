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

func TestRegistrationService_Register_Unit(t *testing.T) {

	t.Run("successful_registration", func(t *testing.T) {
		mockRepo := repository.NewMockUserRepository(t)
		mockEmail := NewMockEmailVerifier(t)
		svc := NewRegistrationService(mockRepo, mockEmail)

		email := "newuser@test.ru"
		password := "password123"

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil).Once()
		mockEmail.On("SendVerification", mock.Anything, mock.Anything, email).Return(nil).Once()

		err := svc.Register(context.Background(), email, password)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockEmail.AssertExpectations(t)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		mockRepo := repository.NewMockUserRepository(t)
		mockEmail := NewMockEmailVerifier(t)
		svc := NewRegistrationService(mockRepo, mockEmail)

		email := "existing@test.ru"
		password := "password123"

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(ErrUserAlreadyExists).Once()

		err := svc.Register(context.Background(), email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserAlreadyExists)
		mockEmail.AssertNotCalled(t, "SendVerification", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("invalid_email", func(t *testing.T) {
		mockRepo := repository.NewMockUserRepository(t)
		mockEmail := NewMockEmailVerifier(t)
		svc := NewRegistrationService(mockRepo, mockEmail)

		email := "not_an_email"
		password := "password123"

		err := svc.Register(context.Background(), email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEmail)
		mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("short_password", func(t *testing.T) {
		mockRepo := repository.NewMockUserRepository(t)
		mockEmail := NewMockEmailVerifier(t)
		svc := NewRegistrationService(mockRepo, mockEmail)

		email := "test@test.ru"
		shortPass := "12345"

		err := svc.Register(context.Background(), email, shortPass)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPasswordTooShort)
		mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("db_error", func(t *testing.T) {
		mockRepo := repository.NewMockUserRepository(t)
		mockEmail := NewMockEmailVerifier(t)
		svc := NewRegistrationService(mockRepo, mockEmail)

		email := "test@test.ru"
		password := "password123"

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).
			Return(errors.New("database connection lost")).Once()

		err := svc.Register(context.Background(), email, password)

		assert.Error(t, err)
		mockEmail.AssertNotCalled(t, "SendVerification", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("email_normalized_to_lowercase", func(t *testing.T) {
		mockRepo := repository.NewMockUserRepository(t)
		mockEmail := NewMockEmailVerifier(t)
		svc := NewRegistrationService(mockRepo, mockEmail)

		email := "UPPERCASE@TEST.RU"
		password := "password123"
		var capturedUser *models.User

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
			capturedUser = user
			return true
		})).Return(nil).Once()

		mockEmail.On("SendVerification", mock.Anything, mock.Anything, "uppercase@test.ru").Return(nil).Once()

		err := svc.Register(context.Background(), email, password)

		assert.NoError(t, err)
		assert.Equal(t, "uppercase@test.ru", capturedUser.Email)
	})
}
