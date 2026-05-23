package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// ─── Connect to PostgreSQL ─────────────────────
func Connect(databaseURL string) *sql.DB {
	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ PostgreSQL not reachable: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")
	return DB
}
