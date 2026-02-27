package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/0xd1sph0l1dus/snakedex/internal/models"
)

// DetailClient calls the detail-service JSON API.
type DetailClient struct {
	baseURL string
	http    *http.Client
}

// NewDetailClient creates a client targeting baseURL (e.g. "http://detail-service:8082").
func NewDetailClient(baseURL string) *DetailClient {
	return &DetailClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// GetDetails fetches full taxon details from the detail-service.
func (c *DetailClient) GetDetails(id int) (*models.SnakeDetails, error) {
	resp, err := c.http.Get(fmt.Sprintf("%s/taxa/%d", c.baseURL, id))
	if err != nil {
		return nil, fmt.Errorf("detail-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("detail-service: taxon %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("detail-service: status %d", resp.StatusCode)
	}

	var details models.SnakeDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("detail-service decode: %w", err)
	}
	return &details, nil
}
