package db

import (
	"database/sql"
	"log"
	"log/slog"

	_ "github.com/lib/pq"
)

var DB *sql.DB

const schema1 = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

const schema2 = `CREATE TABLE IF NOT EXISTS orders (
    number TEXT PRIMARY KEY,      -- Order number as the unique identifier
    username TEXT NOT NULL,       -- Foreign key linking to users table
    status TEXT NOT NULL,         -- Order status
    accrual FLOAT DEFAULT 0,      -- Accrual amount, defaulting to 0
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Timestamp of order creation
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
);
`

func InitDB(dsn string) *sql.DB {
	// Connect to the database
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Verify the connection
	err = DB.Ping()
	if err != nil {
		log.Fatalf("PostgreSQL is not reachable: %v", err)
	}

	slog.Info("Connected to PostgreSQL successfully!")

	_, err = DB.Exec(schema1)
	if err != nil {
		log.Fatalf("error creating schema: %v", err)
	}

	_, err = DB.Exec(schema2)
	if err != nil {
		log.Fatalf("error creating schema: %v", err)
	}
	return DB
}
