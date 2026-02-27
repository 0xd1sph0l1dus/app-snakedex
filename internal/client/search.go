// Package client provides typed HTTP clients for inter-service communication.
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/0xd1sph0l1dus/snakedex/internal/models"
)

// SearchClient calls the search-service JSON API.
type SearchClient struct {
	baseURL string
	http    *http.Client
}

type searchResponse struct {
	Cards []models.SnakeCard `json:"cards"`
	Total int                `json:"total"`
}

// NewSearchClient creates a client targeting baseURL (e.g. "http://search-service:8081").
func NewSearchClient(baseURL string) *SearchClient {
	return &SearchClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// Search queries the search-service and returns matching cards and total count.
func (c *SearchClient) Search(query string, page, minObs int) ([]models.SnakeCard, int, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("page", strconv.Itoa(page))
	if minObs > 0 {
		params.Set("min_obs", strconv.Itoa(minObs))
	}

	resp, err := c.http.Get(fmt.Sprintf("%s/search?%s", c.baseURL, params.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("search-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("search-service: status %d", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("search-service decode: %w", err)
	}
	return result.Cards, result.Total, nil
}
