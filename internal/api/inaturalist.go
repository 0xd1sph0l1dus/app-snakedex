package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/0xd1sph0l1dus/snakedex/internal/models"
)

const (
	baseURL      = "https://api.inaturalist.org/v1"
	serpentaeTaxonID = "85553"
)

var client = &http.Client{Timeout: 10 * time.Second}


const PerPage = 20

type taxaResponse struct {
	TotalResults int     `json:"total_results"`
	Results      []taxon `json:"results"`
}

type taxon struct {
	ID                 int          `json:"id"`
	Name               string       `json:"name"`
	PreferredCommonName string      `json:"preferred_common_name"`
	DefaultPhoto       *taxonPhoto  `json:"default_photo"`
	ObservationsCount  int          `json:"observations_count"`
	WikipediaURL       string       `json:"wikipedia_url"`
	Family             string       `json:"-"` 
	Genus              string       `json:"-"`
	Ancestors          []ancestor   `json:"ancestors"`
	WikipediaSummary   string       `json:"wikipedia_summary"`
}

type taxonPhoto struct {
	MediumURL string `json:"medium_url"`
	SquareURL string `json:"square_url"`
}

type ancestor struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Rank string `json:"rank"`
}

type observationsResponse struct {
	Results []observation `json:"results"`
}

type observation struct {
	PlaceGuess string             `json:"place_guess"`
	Photos     []observationPhoto `json:"photos"`
	GeoJSON    *observationGeoJSON `json:"geojson"`
}

type observationPhoto struct {
	URL string `json:"url"`
}

// GeoJSON Point: coordinates are [longitude, latitude].
type observationGeoJSON struct {
	Coordinates [2]float64 `json:"coordinates"`
}

type observationData struct {
	Locations   []string
	Photos      []string
	Coordinates []models.Coordinate
}

// SearchSnakes queries iNaturalist taxa under Serpentes.
// Returns the cards for the requested page and the total number of results.
// minObs filters by minimum observation count (0 = no filter).
func SearchSnakes(query string, page, minObs int) ([]models.SnakeCard, int, error) {
	if page < 1 {
		page = 1
	}
	params := url.Values{}
	params.Set("q", query)
	params.Set("taxon_id", serpentaeTaxonID)
	params.Set("rank", "species")
	params.Set("per_page", strconv.Itoa(PerPage))
	params.Set("page", strconv.Itoa(page))
	params.Set("locale", "en")
	if minObs > 0 {
		params.Set("min_observations_count", strconv.Itoa(minObs))
	}

	resp, err := client.Get(fmt.Sprintf("%s/taxa?%s", baseURL, params.Encode()))
	if err != nil {
		return nil, 0, fmt.Errorf("inaturalist search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("inaturalist search: status %d", resp.StatusCode)
	}

	var result taxaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("inaturalist search decode: %w", err)
	}

	cards := make([]models.SnakeCard, 0, len(result.Results))
	for _, t := range result.Results {
		cards = append(cards, taxonToCard(t))
	}
	return cards, result.TotalResults, nil
}

// GetSnakeDetails fetches full details for a taxon by ID.
func GetSnakeDetails(id int) (*models.SnakeDetails, error) {
	resp, err := client.Get(fmt.Sprintf("%s/taxa/%d", baseURL, id))
	if err != nil {
		return nil, fmt.Errorf("inaturalist detail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inaturalist detail: status %d", resp.StatusCode)
	}

	// The single-taxon endpoint wraps in a taxaResponse too.
	var result taxaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("inaturalist detail decode: %w", err)
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("taxon %d not found", id)
	}

	t := result.Results[0]
	details := &models.SnakeDetails{
		SnakeCard:        taxonToCard(t),
		WikipediaSummary: t.WikipediaSummary,
	}

	// Extract family and genus from ancestors.
	for _, a := range t.Ancestors {
		switch strings.ToLower(a.Rank) {
		case "family":
			details.Family = a.Name
		case "genus":
			details.Genus = a.Name
		}
	}

	// Fetch observation data (photos, locations, coordinates).
	obsData, err := fetchObservationData(id)
	if err == nil {
		details.RecentLocations = obsData.Locations
		details.Photos = obsData.Photos
		details.Coordinates = obsData.Coordinates
	}

	return details, nil
}

func taxonToCard(t taxon) models.SnakeCard {
	card := models.SnakeCard{
		ID:               t.ID,
		ScientificName:   t.Name,
		CommonName:       t.PreferredCommonName,
		ObservationCount: t.ObservationsCount,
		WikipediaURL:     t.WikipediaURL,
	}
	if t.DefaultPhoto != nil {
		card.PhotoURL = t.DefaultPhoto.MediumURL
		card.ThumbURL = t.DefaultPhoto.SquareURL
	}
	return card
}

func fetchObservationData(taxonID int) (observationData, error) {
	params := url.Values{}
	params.Set("taxon_id", strconv.Itoa(taxonID))
	params.Set("per_page", "20")
	params.Set("quality_grade", "research")
	params.Set("has[]", "photos")

	resp, err := client.Get(fmt.Sprintf("%s/observations?%s", baseURL, params.Encode()))
	if err != nil {
		return observationData{}, fmt.Errorf("inaturalist observations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return observationData{}, fmt.Errorf("inaturalist observations: status %d", resp.StatusCode)
	}

	var result observationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return observationData{}, fmt.Errorf("inaturalist observations decode: %w", err)
	}

	var data observationData
	seenLocs := make(map[string]bool)
	photoCount := 0

	for _, obs := range result.Results {
		if obs.PlaceGuess != "" && !seenLocs[obs.PlaceGuess] {
			seenLocs[obs.PlaceGuess] = true
			data.Locations = append(data.Locations, obs.PlaceGuess)
		}
		if photoCount < 8 && len(obs.Photos) > 0 {
			if u := photoMediumURL(obs.Photos[0].URL); u != "" {
				data.Photos = append(data.Photos, u)
				photoCount++
			}
		}
		if obs.GeoJSON != nil {
			lng, lat := obs.GeoJSON.Coordinates[0], obs.GeoJSON.Coordinates[1]
			if lat != 0 || lng != 0 {
				data.Coordinates = append(data.Coordinates, models.Coordinate{Lat: lat, Lng: lng})
			}
		}
	}
	return data, nil
}

func photoMediumURL(u string) string {
	return strings.Replace(u, "/square.", "/medium.", 1)
}
