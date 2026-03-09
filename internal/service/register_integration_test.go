package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationService_Register_Integration(t *testing.T) {
	_, err := testDB.Exec("TRUNCATE users CASCADE")
	require.NoError(t, err)

	t.Run("success_registration", func(t *testing.T) {
		ctx := context.Background()
		email := "newuser@test.com"
		password := "password123"

		err := testRegSvc.Register(ctx, email, password)

		assert.NoError(t, err)
	})

	t.Run("register_and_verify_email", func(t *testing.T) {
		ctx := context.Background()
		email := "verify-me@example.com"

		err := testRegSvc.Register(ctx, email, "strong-pass-123")
		assert.NoError(t, err)

		tokenStr := fetchTokenFromMailpit(t, testMailpitAPI)
		parsedUUID, err := uuid.Parse(tokenStr)
		require.NoError(t, err, "Need valid UUID")

		err = verifySvc.Verify(ctx, parsedUUID)
		assert.NoError(t, err)

		var isVerified bool
		err = testDB.QueryRow("SELECT is_verified FROM users WHERE email = $1", email).Scan(&isVerified)
		assert.NoError(t, err)
		assert.True(t, isVerified)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		ctx := context.Background()
		email := "duplicate@test.com"
		password := "password123"

		err := testRegSvc.Register(ctx, email, password)
		require.NoError(t, err)

		err = testRegSvc.Register(ctx, email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserAlreadyExists)
	})

	t.Run("invalid_email_format", func(t *testing.T) {
		ctx := context.Background()
		email := "not-an-email"
		password := "password123"

		err := testRegSvc.Register(ctx, email, password)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidEmail)
	})

	t.Run("short_password", func(t *testing.T) {
		ctx := context.Background()
		email := "shortpass@test.com"
		shortPass := "1234567"

		err := testRegSvc.Register(ctx, email, shortPass)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPasswordTooShort)
	})

	t.Run("sql_injection_attempt", func(t *testing.T) {
		ctx := context.Background()
		maliciousEmail := "'; DROP TABLE users; --"
		password := "password123"

		err := testRegSvc.Register(ctx, maliciousEmail, password)

		assert.Error(t, err)
	})

	t.Run("email_normalized_to_lowercase", func(t *testing.T) {
		ctx := context.Background()
		email := "UPPERCASE@TEST.COM"
		password := "password123"

		err := testRegSvc.Register(ctx, email, password)

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

		err := testRegSvc.Register(ctx, email, password)
		require.NoError(t, err)

		token, err := testAuthSvc.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}
