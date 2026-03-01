package repository

import (
	"context"
	"database/sql"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type Repository struct {
	Users UserRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Users: &UserPostgres{db: db},
	}
}
