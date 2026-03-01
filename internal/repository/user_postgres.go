package repository

import (
	"context"
	"database/sql"

	"github.com/dvprokofiev/seating-generator-api/internal/models"
)

type UserPostgres struct {
	db *sql.DB
}

func (r *UserPostgres) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`
	return r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash).Scan(&user.ID)
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
