// Package service defines the contracts between the frontend and backend services.
// Both the direct API calls (monolith) and HTTP clients (microservices) implement
// these interfaces, making them interchangeable.
package service

import "github.com/0xd1sph0l1dus/snakedex/internal/models"

// Searcher returns paginated snake cards from a text query.
type Searcher interface {
	Search(query string, page, minObs int) ([]models.SnakeCard, int, error)
}

// Detailer returns full details for a single taxon by ID.
type Detailer interface {
	GetDetails(id int) (*models.SnakeDetails, error)
}
