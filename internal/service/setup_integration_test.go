package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/dvprokofiev/seating-generator-api/internal/database"
	"github.com/dvprokofiev/seating-generator-api/internal/mailer"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	verifySvc      EmailVerifier
	testMailpitAPI string
)

func setupMailpit(ctx context.Context) (string, string, testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "axllent/mailpit",
			ExposedPorts: []string{"1025/tcp", "8025/tcp"},
			WaitingFor:   wait.ForHTTP("/").WithPort("8025/tcp"),
		},
		Started: true,
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to start mailpit: %w", err)
	}

	host, _ := container.Host(ctx)
	smtpPort, _ := container.MappedPort(ctx, "1025")
	apiPort, _ := container.MappedPort(ctx, "8025")

	return fmt.Sprintf("%s:%s", host, smtpPort.Port()), fmt.Sprintf("%s:%s", host, apiPort.Port()), container, nil
}

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

	smtpAddr, apiAddr, mpContainer, err := setupMailpit(ctx)

	if err != nil {
		panic(err)
	}

	testMailpitAPI = apiAddr

	repo := repository.NewRepository(testDB)

	host, port, _ := net.SplitHostPort(smtpAddr)
	realMailer := mailer.NewSMTPMailer(host, port, "", "", "noreply@test.ru", "http://localhost:8080")

	verifySvc = NewEmailVerification(
		repo.EmailVerification,
		repo.Users,
		realMailer,
		EmailVerificationConfig{TTL: 24 * time.Hour},
	)
	testAuthSvc = NewAuthService(repo.Users, "test-secret")
	testRegSvc = NewRegistrationService(repo.Users, verifySvc)

	code := m.Run()

	testDB.Close()
	pgContainer.Terminate(ctx)
	mpContainer.Terminate(ctx)

	os.Exit(code)
}

func fetchTokenFromMailpit(t *testing.T, apiAddr string) string {
	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/messages", apiAddr))
	require.NoError(t, err)
	defer resp.Body.Close()

	var data struct {
		Messages []struct{ ID string } `json:"messages"`
	}
	json.NewDecoder(resp.Body).Decode(&data)
	require.NotEmpty(t, data.Messages, "No messages in Mailpit")

	msgID := data.Messages[0].ID
	resp, _ = http.Get(fmt.Sprintf("http://%s/api/v1/message/%s", apiAddr, msgID))
	var msg struct{ HTML string }
	json.NewDecoder(resp.Body).Decode(&msg)

	re := regexp.MustCompile(`token=([a-fA-F0-9-]{36})`)
	matches := re.FindStringSubmatch(msg.HTML)
	require.Len(t, matches, 2, "Token not found in email body")

	return matches[1]
}
