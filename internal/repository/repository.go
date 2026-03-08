package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
	"github.com/google/uuid"
)

//go:generate mockery --name=User --inpackage --case=snake
//go:generate mockery --name=Email --inpackage --case=snake

type User interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateVerified(ctx context.Context, userID uuid.UUID, isVerified bool) error
}

type Email interface {
	CreateOrUpdate(ctx context.Context, userID uuid.UUID, ttl time.Duration) (uuid.UUID, error)
	VerifyAndConsume(ctx context.Context, token uuid.UUID) (uuid.UUID, error)
	CleanupExpired(ctx context.Context) (int64, error)
}

type Repository struct {
	Users             User
	EmailVerification Email
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Users:             &UserPostgres{db: db},
		EmailVerification: &EmailVerificationPostgres{db: db},
	}
}
