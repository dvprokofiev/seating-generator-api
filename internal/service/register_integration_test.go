package service

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Register_Integration(t *testing.T) {
	_, err := testDB.Exec("TRUNCATE users CASCADE")
	require.NoError(t, err)

	t.Run("success_registration_db", func(t *testing.T) {
		ctx := context.Background()
		email := "newuser@test.com"
		password := "password123"

		err := testSvc.Register(ctx, email, password)

		assert.NoError(t, err)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		ctx := context.Background()
		email := "duplicate@test.com"
		password := "password123"

		err := testSvc.Register(ctx, email, password)
		require.NoError(t, err)

		err = testSvc.Register(ctx, email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserAlreadyExists)
	})

	t.Run("invalid_email_format", func(t *testing.T) {
		ctx := context.Background()
		email := "not-an-email"
		password := "password123"

		err := testSvc.Register(ctx, email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEmail)
	})

	t.Run("short_password", func(t *testing.T) {
		ctx := context.Background()
		email := "shortpass@test.com"
		shortPass := "1234567"

		err := testSvc.Register(ctx, email, shortPass)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPasswordTooShort)
	})

	t.Run("sql_injection_attempt", func(t *testing.T) {
		ctx := context.Background()
		maliciousEmail := "'; DROP TABLE users; --"
		password := "password123"

		err := testSvc.Register(ctx, maliciousEmail, password)

		assert.Error(t, err)
	})

	t.Run("email_normalized_to_lowercase", func(t *testing.T) {
		ctx := context.Background()
		email := "UPPERCASE@TEST.COM"
		password := "password123"

		err := testSvc.Register(ctx, email, password)

		assert.NoError(t, err)

		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE email = 'uppercase@test.com'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("user_can_login_after_registration", func(t *testing.T) {
		ctx := context.Background()
		email := "loginafter@test.com"
		password := "password123"

		err := testSvc.Register(ctx, email, password)
		require.NoError(t, err)

		token, err := testSvc.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}
