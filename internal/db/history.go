package db

import (
	"fmt"
	"time"
)

type SearchEntry struct {
	Query      string
	Count      int
	SearchedAt time.Time
}

// AddSearch logs a search query and its result count.
func AddSearch(query string, count int) error {
	_, err := DB.Exec(
		`INSERT INTO search_history (query, results_count) VALUES (?, ?)`,
		query, count,
	)
	return err
}

// ListHistory returns the last n search entries.
func ListHistory(limit int) ([]SearchEntry, error) {
	rows, err := DB.Query(
		`SELECT query, results_count, searched_at FROM search_history
		 ORDER BY searched_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}
	defer rows.Close()

	var entries []SearchEntry
	for rows.Next() {
		var e SearchEntry
		if err := rows.Scan(&e.Query, &e.Count, &e.SearchedAt); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
