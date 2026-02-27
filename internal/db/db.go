package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS favorites (
    taxon_id        INTEGER PRIMARY KEY,
    scientific_name TEXT    NOT NULL,
    common_name     TEXT    NOT NULL DEFAULT '',
    photo_url       TEXT    NOT NULL DEFAULT '',
    thumb_url       TEXT    NOT NULL DEFAULT '',
    saved_at        DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS search_history (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    query         TEXT    NOT NULL,
    results_count INTEGER NOT NULL DEFAULT 0,
    searched_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

// Init opens the SQLite database at path and runs migrations.
func Init(path string) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("db open: %w", err)
	}
	// SQLite supports a single writer; cap connections to avoid locking errors.
	db.SetMaxOpenConns(1)
	DB = db
	return migrate()
}

func migrate() error {
	if _, err := DB.Exec(schema); err != nil {
		return fmt.Errorf("db migrate: %w", err)
	}
	return nil
}
