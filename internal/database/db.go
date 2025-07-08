package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func Initialize(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	tables := []string{
		// Deployment tokens table
		`CREATE TABLE IF NOT EXISTS deployment_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token TEXT UNIQUE NOT NULL,
			valid_until DATETIME NOT NULL,
			max_uses INTEGER NOT NULL,
			used_count INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL,
			notes TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_token ON deployment_tokens(token);`,

		// Agents table
		`CREATE TABLE IF NOT EXISTS agents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id TEXT UNIQUE NOT NULL,
			api_key TEXT UNIQUE NOT NULL,
			hostname TEXT,
			username TEXT,
			os TEXT,
			arch TEXT,
			activated_at DATETIME NOT NULL,
			last_seen DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_agent_id ON agents(agent_id);
		CREATE INDEX IF NOT EXISTS idx_api_key ON agents(api_key);`,

		// Command history table
		`CREATE TABLE IF NOT EXISTS command_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id TEXT NOT NULL,
			command TEXT,
			response TEXT,
			executed_at DATETIME NOT NULL,
			FOREIGN KEY(agent_id) REFERENCES agents(agent_id)
		);
		CREATE INDEX IF NOT EXISTS idx_cmd_agent_id ON command_history(agent_id);`,

		// Token usage tracking table
		`CREATE TABLE IF NOT EXISTS token_agent_usage (
			token_id INTEGER NOT NULL,
			agent_id TEXT NOT NULL,
			used_at DATETIME NOT NULL,
			PRIMARY KEY (token_id, agent_id),
			FOREIGN KEY(token_id) REFERENCES deployment_tokens(id),
			FOREIGN KEY(agent_id) REFERENCES agents(agent_id)
		);`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return err
		}
	}

	return nil
}
