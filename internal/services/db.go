package services

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(dbURL string) (*sql.DB, error) {
	if dbURL == "" {
		log.Println("Warning: DATABASE_URL is not set. Governance decisions will not be persisted.")
		return nil, nil // Return nil so agents can selectively skip DB operations
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Initialize tables
	query := `
	CREATE TABLE IF NOT EXISTS governance_decisions (
		id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		pipeline_id TEXT NOT NULL,
		risk_level  TEXT NOT NULL,
		patch_hash  TEXT NOT NULL,
		decision    TEXT NOT NULL,
		llm_reason  TEXT,
		decided_at  TIMESTAMPTZ DEFAULT now()
	);`

	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database logic initialized routing ready.")
	return db, nil
}
