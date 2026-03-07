package repository

import (
	"context"
	"database/sql"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
	"github.com/google/uuid"
)

//go:generate mockery --name=UserRepository --inpackage --case=snake

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateVerified(ctx context.Context, userID uuid.UUID, isVerified bool) error
}

type Repository struct {
	Users UserRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Users: &UserPostgres{db: db},
	}
}
