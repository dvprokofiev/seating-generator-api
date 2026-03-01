package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/dvprokofiev/seating-generator-api/internal/database"
	"github.com/dvprokofiev/seating-generator-api/internal/handler"
	"github.com/dvprokofiev/seating-generator-api/internal/repository"
	"github.com/dvprokofiev/seating-generator-api/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No required environment variables found, aborting now")
		return
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	db, err := database.InitDB(dbUser, dbPass, dbHost, dbPort, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repos := repository.NewRepository(db)
	authService := service.NewAuthService(repos.Users, os.Getenv("JWT_SECRET"))
	authHandler := handler.NewAuthHandler(authService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", authHandler.Login)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
