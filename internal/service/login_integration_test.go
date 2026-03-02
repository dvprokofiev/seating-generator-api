package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dvprokofiev/seating-generator-api/internal/database"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
)

var (
	testDB  *sql.DB
	testSvc AuthService
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

	migrationsPath, _ := filepath.Abs("../database/migrations")
	err = database.RunMigrations(testDB)
	fmt.Printf("Running migrations from: %s\n", migrationsPath)

	repo := repository.NewRepository(testDB)
	testSvc = NewAuthService(repo.Users, "test-secret")

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

		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err := testDB.Exec("INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)",
			uuid.New(), email, string(hash))
		require.NoError(t, err)

		token, err := testSvc.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("fail_login_wrong_password", func(t *testing.T) {
		ctx := context.Background()
		email := "wa@test.com"
		password := "correct-password"

		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err := testDB.Exec("INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)",
			uuid.New(), email, string(hash))
		require.NoError(t, err)

		token, err := testSvc.Login(ctx, email, "incorrect-password")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})
	t.Run("user_not_found", func(t *testing.T) {
		ctx := context.Background()
		token, err := testSvc.Login(ctx, "no-such-user@test.com", "password1234")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("sql_injection_attempt", func(t *testing.T) {
		ctx := context.Background()
		maliciousEmail := "' OR 1=1; --"
		token, err := testSvc.Login(ctx, maliciousEmail, "any")

		assert.Error(t, err)
		assert.Empty(t, token)
	})

	t.Run("case_sensitive_email", func(t *testing.T) {
		ctx := context.Background()
		email := "User@Example.com"
		pass := "pass1234"
		hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

		_, err := testDB.Exec("INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)",
			uuid.New(), "user@example.com", string(hash))
		require.NoError(t, err)

		token, err := testSvc.Login(ctx, email, pass)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}
