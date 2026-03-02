package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
)

var ErrDuplicateEmail = errors.New("Email address is already in use")

type UserPostgres struct {
	db *sql.DB
}

func (r *UserPostgres) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, email, password_hash, created_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.PasswordHash, user.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "23505") {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *UserPostgres) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	query := `SELECT id, email, password_hash FROM users WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash)

	if err != nil {
		return nil, err
	}
	return &u, nil
}
