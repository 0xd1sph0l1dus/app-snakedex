package models

type SnakeCard struct {
	ID               int
	ScientificName   string
	CommonName       string
	PhotoURL         string
	ThumbURL         string
	ObservationCount int
	WikipediaURL     string
}

type Coordinate struct {
	Lat float64
	Lng float64
}

type SnakeDetails struct {
	SnakeCard
	Family           string
	Genus            string
	WikipediaSummary string
	RecentLocations  []string
	Photos           []string
	Coordinates      []Coordinate
}
