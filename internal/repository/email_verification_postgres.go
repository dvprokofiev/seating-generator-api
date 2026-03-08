package repository

import (
	context "context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTokenNotFound = errors.New("Verification token was not found")
	ErrTokenExpired  = errors.New("Verification token expired")
)

type EmailVerificationPostgres struct {
	db *sql.DB
}

func NewEmailVerificationRepo(db *sql.DB) *EmailVerificationPostgres {
	return &EmailVerificationPostgres{db: db}
}

func (r *EmailVerificationPostgres) CreateOrUpdate(ctx context.Context, userID uuid.UUID, ttl time.Duration) (uuid.UUID, error) {
	newToken := uuid.New()
	expiresAt := time.Now().Add(ttl)

	query := `
        INSERT INTO email_verifications (user_id, token, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            token = EXCLUDED.token,
            expires_at = EXCLUDED.expires_at,
            created_at = CURRENT_TIMESTAMP
        RETURNING token;`

	var token uuid.UUID
	err := r.db.QueryRowContext(ctx, query, userID, newToken, expiresAt).Scan(&token)
	return token, err
}

func (r *EmailVerificationPostgres) VerifyAndConsume(ctx context.Context, token uuid.UUID) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	var userID uuid.UUID
	var expiresAt time.Time

	query := `SELECT user_id, expires_at FROM email_verifications WHERE token = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, query, token).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrTokenNotFound
		}
		return uuid.Nil, err
	}

	if time.Now().After(expiresAt) {
		_, _ = tx.ExecContext(ctx, "DELETE FROM email_verifications WHERE token = $1", token)
		_ = tx.Commit()
		return uuid.Nil, ErrTokenExpired
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM email_verifications WHERE token = $1", token)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, tx.Commit()
}

func (r *EmailVerificationPostgres) CleanupExpired(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, "DELETE FROM email_verifications WHERE expires_at < $1", time.Now())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
