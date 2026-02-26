package handlers

import (
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/0xd1sph0l1dus/snakedex/internal/api"
	"github.com/0xd1sph0l1dus/snakedex/internal/models"
)

var tmpl *template.Template

func Init(t *template.Template) {
	tmpl = t
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// Index handles GET /
func Index(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		log.Printf("index template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Search handles GET /search?q=...&page=...
func Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		if isHTMX(r) {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	minObs, _ := strconv.Atoi(r.URL.Query().Get("min_obs"))
	sortOrder := r.URL.Query().Get("sort") // "obs_asc" | "obs_desc" | ""

	cards, total, err := api.SearchSnakes(query, page, minObs)
	if err != nil {
		log.Printf("search error: %v", err)
		http.Error(w, "search failed", http.StatusBadGateway)
		return
	}

	sortCards(cards, sortOrder)

	totalPages := (total + api.PerPage - 1) / api.PerPage
	if totalPages < 1 {
		totalPages = 1
	}

	data := map[string]interface{}{
		"Query":      query,
		"Cards":      cards,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
		"HasPrev":    page > 1,
		"HasNext":    page < totalPages,
		"PrevPage":   page - 1,
		"NextPage":   page + 1,
		"MinObs":     minObs,
		"Sort":       sortOrder,
	}

	if isHTMX(r) {
		if err := tmpl.ExecuteTemplate(w, "results.html", data); err != nil {
			log.Printf("results template: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("index+results template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// SnakeDetail handles GET /snake/{id}
func SnakeDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	details, err := api.GetSnakeDetails(id)
	if err != nil {
		log.Printf("detail error id=%d: %v", id, err)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if isHTMX(r) {
		if err := tmpl.ExecuteTemplate(w, "detail.html", details); err != nil {
			log.Printf("detail template: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	if err := tmpl.ExecuteTemplate(w, "snake.html", details); err != nil {
		log.Printf("snake template: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func sortCards(cards []models.SnakeCard, order string) {
	switch order {
	case "obs_desc":
		sort.Slice(cards, func(i, j int) bool {
			return cards[i].ObservationCount > cards[j].ObservationCount
		})
	case "obs_asc":
		sort.Slice(cards, func(i, j int) bool {
			return cards[i].ObservationCount < cards[j].ObservationCount
		})
	}
}
