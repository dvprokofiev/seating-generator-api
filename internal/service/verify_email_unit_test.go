package service

import (
	"context"
	"testing"
	"time"

	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEmailVerification_Verify(t *testing.T) {
	ctx := context.Background()

	t.Run("success_verification", func(t *testing.T) {
		mockEmailRepo := repository.NewMockEmail(t)
		mockUserRepo := repository.NewMockUserRepository(t)
		svc := NewEmailVerification(mockEmailRepo, mockUserRepo, nil, EmailVerificationConfig{TTL: 24 * time.Hour})

		token := uuid.New()
		userID := uuid.New()

		mockEmailRepo.On("VerifyAndConsume", ctx, token).Return(userID, nil).Once()

		mockUserRepo.On("UpdateVerified", ctx, userID, true).Return(nil).Once()

		err := svc.Verify(ctx, token)
		assert.NoError(t, err)
	})

	t.Run("token_not_exists", func(t *testing.T) {
		mockEmailRepo := repository.NewMockEmail(t)
		mockUserRepo := repository.NewMockUserRepository(t)
		svc := NewEmailVerification(mockEmailRepo, mockUserRepo, nil, EmailVerificationConfig{TTL: 24 * time.Hour})

		token := uuid.New()

		mockEmailRepo.On("VerifyAndConsume", ctx, token).Return(uuid.Nil, repository.ErrTokenNotFound).Once()

		err := svc.Verify(ctx, token)

		assert.ErrorIs(t, err, ErrTokenNotFound)
		mockUserRepo.AssertNotCalled(t, "UpdateVerified", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("token_expired", func(t *testing.T) {
		mockEmailRepo := repository.NewMockEmail(t)
		mockUserRepo := repository.NewMockUserRepository(t)
		svc := NewEmailVerification(mockEmailRepo, mockUserRepo, nil, EmailVerificationConfig{TTL: 24 * time.Hour})

		token := uuid.New()

		mockEmailRepo.On("VerifyAndConsume", ctx, token).Return(uuid.Nil, repository.ErrTokenExpired).Once()

		err := svc.Verify(ctx, token)

		assert.ErrorIs(t, err, ErrTokenExpired)
		mockUserRepo.AssertNotCalled(t, "UpdateVerified", mock.Anything, mock.Anything, mock.Anything)
	})
}
