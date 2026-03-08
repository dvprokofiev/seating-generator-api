package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dvprokofiev/seating-generator-api/internal/database"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB      *sql.DB
	testAuthSvc AuthService
	testRegSvc  RegistrationService
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(15*time.Second)),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to start container: %s", err))
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	err = database.RunMigrations(testDB)
	if err != nil {
		panic(err)
	}

	repo := repository.NewRepository(testDB)

	mockMail := &MockEmailVerifier{}
	mockMail.On("SendVerification", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	testAuthSvc = NewAuthService(repo.Users, "test-secret")
	testRegSvc = NewRegistrationService(repo.Users, mockMail)

	code := m.Run()

	testDB.Close()
	pgContainer.Terminate(ctx)

	os.Exit(code)
}

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
