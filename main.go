package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"pets_project/internal/db"
	"pets_project/internal/handlers"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (if exists)
	if err := godotenv.Load(); err != nil {
		log.Println("WARN: .env file not found, using system environment variables")
	}

	// Initialize custom logger
	handlers.InitLogger()
	handlers.Info("Starting PETS_PROJECT backend initialization")

	// Load DB env variables
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	// Build PostgreSQL connection string
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	// Initialize DB connection
	dbConn := db.InitDB(connStr)
	defer dbConn.Close()
	handlers.Info("Database connection established successfully")

	// Shared environment instance
	env := &handlers.Env{DB: dbConn}

	// ============================================================
	// PROTECTED ROUTER (JWT REQUIRED)
	// ============================================================

	apiRouter := http.NewServeMux()

	// Pets routes
	apiRouter.HandleFunc("/pets", env.PetsHandler)
	apiRouter.HandleFunc("/pets/", env.PetsHandler)

	// Owners routes
	apiRouter.HandleFunc("/owners", env.OwnersHandler)
	apiRouter.HandleFunc("/owners/", env.OwnersHandler)

	// Appointments routes
	apiRouter.HandleFunc("/appointments", env.AppointmentsHandler)
	apiRouter.HandleFunc("/appointments/", env.AppointmentsHandler)

	// File upload & download
	apiRouter.HandleFunc("/upload", env.UploadFileHandler)
	apiRouter.HandleFunc("/download", env.DownloadFileHandler)

	// File list & delete
	apiRouter.HandleFunc("/files", env.ListFilesHandler)
	apiRouter.HandleFunc("/files/delete", env.DeleteFileHandler)

	handlers.Info("All protected routes registered successfully")

	// Wrap with JWT middleware
	protectedAPI := env.JwtAuthMiddleware(apiRouter)

	// ============================================================
	// PUBLIC ROUTER (NO AUTH REQUIRED)
	// ============================================================

	masterRouter := http.NewServeMux()

	// Public endpoints
	masterRouter.HandleFunc("/signup", env.SignupHandler)
	masterRouter.HandleFunc("/login", env.LoginHandler)

	// All other endpoints require JWT
	masterRouter.Handle("/", protectedAPI)

	// ============================================================
	// START SERVER
	// ============================================================

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
		handlers.Warn("SERVER_PORT not set — using default :8081")
	}

	handlers.Info("Server running on port :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, masterRouter))
}
