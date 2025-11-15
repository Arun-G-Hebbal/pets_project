package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// InitDB initializes and returns a database connection
func InitDB(connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("ERROR: Failed to open database connection: %v", err)
	}

	// Verify connection
	if err = db.Ping(); err != nil {
		log.Fatalf("ERROR: Could not connect to database: %v", err)
	}

	fmt.Println("INFO: Successfully connected to the database.")
	return db
}
