package service

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB      *sql.DB
	testAuthSvc AuthService
	testRegSvc  RegistrationService
)

func TestAuthService_Login_Integration(t *testing.T) {
	_, err := testDB.Exec("TRUNCATE users CASCADE")
	require.NoError(t, err)

	t.Run("success_login_db", func(t *testing.T) {
		ctx := context.Background()
		email := "real-user@test.com"
		password := "password123"

		err := testRegSvc.Register(ctx, email, password)
		require.NoError(t, err)

		token, err := testAuthSvc.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("fail_login_wrong_password", func(t *testing.T) {
		ctx := context.Background()
		email := "wa@test.com"
		password := "correct-password"

		err := testRegSvc.Register(ctx, email, password)
		require.NoError(t, err)

		token, err := testAuthSvc.Login(ctx, email, "incorrect-password")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("user_not_found", func(t *testing.T) {
		ctx := context.Background()
		token, err := testAuthSvc.Login(ctx, "no-such-user@test.com", "password1234")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("case_sensitive_email", func(t *testing.T) {
		ctx := context.Background()
		email := "User@Example.com"
		pass := "pass1234"

		err := testRegSvc.Register(ctx, email, pass)
		require.NoError(t, err)

		token, err := testAuthSvc.Login(ctx, email, pass)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}
