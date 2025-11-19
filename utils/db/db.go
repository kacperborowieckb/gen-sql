package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// NewConnection creates and returns a new *sql.DB connection pool.
// It should be called ONCE when a service starts.
func NewConnection(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
		return nil, err
	}

	log.Println("Database connection pool established")
	return db, nil
}
