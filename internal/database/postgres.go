package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var ef embed.FS

func getDSN(user, pass, host, port, name string) (dsn string) {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user, pass, host, port, name)
}

func dbConnect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("Error trying to open database: %w", err)
	}

	// set connection pool (for performance)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(ef)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("Error setting up Goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("Error while running migrations: %w", err)
	}

	log.Println("Database is ready, all migrations applied successfully")
	return nil
}

func InitDB(user, pass, host, port, name string) (*sql.DB, error) {
	dsn := getDSN(user, pass, host, port, name)
	db, err := dbConnect(dsn)
	if err != nil {
		return nil, err
	}

	err = RunMigrations(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
