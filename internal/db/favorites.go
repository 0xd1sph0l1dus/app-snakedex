package db

import (
	"fmt"

	"github.com/0xd1sph0l1dus/snakedex/internal/models"
)

// AddFavorite inserts or replaces a snake in the favorites table.
func AddFavorite(s models.SnakeCard) error {
	_, err := DB.Exec(
		`INSERT OR REPLACE INTO favorites (taxon_id, scientific_name, common_name, photo_url, thumb_url)
		 VALUES (?, ?, ?, ?, ?)`,
		s.ID, s.ScientificName, s.CommonName, s.PhotoURL, s.ThumbURL,
	)
	return err
}

// RemoveFavorite deletes a snake from the favorites table.
func RemoveFavorite(taxonID int) error {
	_, err := DB.Exec(`DELETE FROM favorites WHERE taxon_id = ?`, taxonID)
	return err
}

// IsFavorite reports whether a taxon is saved as a favorite.
func IsFavorite(taxonID int) (bool, error) {
	var count int
	err := DB.QueryRow(`SELECT COUNT(*) FROM favorites WHERE taxon_id = ?`, taxonID).Scan(&count)
	return count > 0, err
}

// ListFavorites returns all favorites ordered by most recently saved.
func ListFavorites() ([]models.SnakeCard, error) {
	rows, err := DB.Query(
		`SELECT taxon_id, scientific_name, common_name, photo_url, thumb_url
		 FROM favorites ORDER BY saved_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()

	var cards []models.SnakeCard
	for rows.Next() {
		var c models.SnakeCard
		if err := rows.Scan(&c.ID, &c.ScientificName, &c.CommonName, &c.PhotoURL, &c.ThumbURL); err != nil {
			return nil, fmt.Errorf("scan favorite: %w", err)
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}
