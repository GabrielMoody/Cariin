package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func Init() *sql.DB {
	connStr := "postgres://postgres:1@localhost:5432/cariin?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	return db
}
