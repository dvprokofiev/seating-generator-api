package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockery --name=RegistrationService --inpackage --case=snake

type RegistrationService interface {
	Register(ctx context.Context, email, password string) error
}

type registrationService struct {
	userRepo     repository.User
	emailService EmailVerifier
}

func NewRegistrationService(uRepo repository.User, eSvc EmailVerifier) RegistrationService {
	return &registrationService{
		userRepo:     uRepo,
		emailService: eSvc,
	}
}

func (s *registrationService) Register(ctx context.Context, email, password string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("Failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        strings.ToLower(email),
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now().UTC(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return ErrUserAlreadyExists
		}
		return err
	}

	err = s.emailService.SendVerification(ctx, user.ID, user.Email)
	if err != nil {
		log.Printf("Failed to send verification email: %v", err)
	}
	return nil
}
