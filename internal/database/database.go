package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" 
)

type DB struct {
	*sql.DB
}

// New creates a new database connection pool
func New(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)                 
	db.SetMaxIdleConns(5)                  
	db.SetConnMaxLifetime(5 * time.Minute) 

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established")

	return &DB{db}, nil
}

func (db *DB) Close() error {
	log.Println("Closing database connection...")
	return db.DB.Close()
}


// RunMigrations executes the database migrations
func (db *DB) RunMigrations() error {
	log.Println("Running database migrations...")

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS generation_requests (
			id SERIAL PRIMARY KEY,
			scenario TEXT NOT NULL,
			row_count INTEGER NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			generated_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create generation_requests table: %w", err)
	}


	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS mock_datasets (
			id SERIAL PRIMARY KEY,
			request_id INTEGER NOT NULL REFERENCES generation_requests(id) ON DELETE CASCADE,
			data JSONB NOT NULL,
			field_names TEXT[] NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create mock_datasets table: %w", err)
	}

	// Create index on request_id for faster lookups
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_mock_datasets_request_id
		ON mock_datasets(request_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Create updated_at trigger function
	_, err = db.Exec(`
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`)
	if err != nil {
		return fmt.Errorf("failed to create trigger function: %w", err)
	}

	// Create trigger for generation_requests
	_, err = db.Exec(`
		DROP TRIGGER IF EXISTS update_generation_requests_updated_at ON generation_requests;
		CREATE TRIGGER update_generation_requests_updated_at
		BEFORE UPDATE ON generation_requests
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();
	`)
	if err != nil {
		return fmt.Errorf("failed to create trigger: %w", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}
