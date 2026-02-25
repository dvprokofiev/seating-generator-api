package main

import (
	"log"
	"os"

	"github.com/dvprokofiev/seating-generator-api/internal/database"
	_ "github.com/jackc/pgx/v5"
)

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	db, err := database.InitDB(dbUser, dbPass, dbHost, dbPort, dbName)
	if err != nil {
		log.Println(err)
		return
	}
}
