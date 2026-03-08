package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dvprokofiev/seating-generator-api/internal/mailer"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/google/uuid"
)

//go:generate mockery --name=EmailVerifier --inpackage --case=snake

var (
	ErrTokenNotFound = errors.New("Verification token was not found")
	ErrTokenExpired  = errors.New("Verification token expired")
)

type EmailVerificationConfig struct {
	TTL time.Duration
}
type EmailVerification struct {
	emailRepo repository.Email
	userRepo  repository.User
	mailer    mailer.Mailer
	cfg       EmailVerificationConfig
}

type EmailVerifier interface {
	SendVerification(ctx context.Context, userID uuid.UUID, email string) error
	Verify(ctx context.Context, token uuid.UUID) error
}

func NewEmailVerification(eRepo repository.Email, uRepo repository.User, m mailer.Mailer, cfg EmailVerificationConfig) EmailVerifier {
	return &EmailVerification{
		emailRepo: eRepo,
		userRepo:  uRepo,
		mailer:    m,
		cfg:       cfg,
	}
}

func (s *EmailVerification) SendVerification(ctx context.Context, userID uuid.UUID, email string) error {
	token, err := s.emailRepo.CreateOrUpdate(ctx, userID, s.cfg.TTL)
	if err != nil {
		return err
	}

	return s.mailer.SendVerificationEmail(email, token)
}

func (s *EmailVerification) Verify(ctx context.Context, token uuid.UUID) error {
	userID, err := s.emailRepo.VerifyAndConsume(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return ErrTokenNotFound
		} else if errors.Is(err, repository.ErrTokenExpired) {
			return ErrTokenExpired
		} else {
			return err
		}
	}

	err = s.userRepo.UpdateVerified(ctx, userID, true)
	if err != nil {
		return fmt.Errorf("Failed to update user status: %w", err)
	}
	return nil
}
