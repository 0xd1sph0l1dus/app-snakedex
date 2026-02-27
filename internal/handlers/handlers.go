package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sort"
	"strconv"

	"github.com/0xd1sph0l1dus/snakedex/internal/api"
	"github.com/0xd1sph0l1dus/snakedex/internal/db"
	"github.com/0xd1sph0l1dus/snakedex/internal/models"
	"github.com/0xd1sph0l1dus/snakedex/internal/service"
)

var (
	tmpl     *template.Template
	searcher service.Searcher
	detailer service.Detailer
)

func Init(t *template.Template, s service.Searcher, d service.Detailer) {
	tmpl = t
	searcher = s
	detailer = d
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

type detailData struct {
	*models.SnakeDetails
	IsFavorited bool
}

// Index handles GET /
func Index(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		slog.Error("index template", "err", err)
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
	sortOrder := r.URL.Query().Get("sort")

	cards, total, err := searcher.Search(query, page, minObs)
	if err != nil {
		slog.Error("search failed", "err", err, "query", query)
		http.Error(w, "search failed", http.StatusBadGateway)
		return
	}

	sortCards(cards, sortOrder)

	go func() {
		if err := db.AddSearch(query, total); err != nil {
			slog.Warn("history log failed", "err", err)
		}
	}()

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
			slog.Error("results template", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Error("index+results template", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// SnakeDetail handles GET /snake/{id}
func SnakeDetail(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	details, err := detailer.GetDetails(id)
	if err != nil {
		slog.Error("detail fetch failed", "err", err, "id", id)
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	isFav, _ := db.IsFavorite(id)
	data := detailData{details, isFav}

	if isHTMX(r) {
		if err := tmpl.ExecuteTemplate(w, "detail.html", data); err != nil {
			slog.Error("detail template", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	if err := tmpl.ExecuteTemplate(w, "snake.html", data); err != nil {
		slog.Error("snake template", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// ToggleFavorite handles POST /snake/{id}/favorite
func ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	isFav, err := db.IsFavorite(id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if isFav {
		if err := db.RemoveFavorite(id); err != nil {
			slog.Error("remove favorite", "err", err, "id", id)
		}
	} else {
		details, err := detailer.GetDetails(id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err := db.AddFavorite(details.SnakeCard); err != nil {
			slog.Error("add favorite", "err", err, "id", id)
		}
	}

	if err := tmpl.ExecuteTemplate(w, "fav-btn.html", map[string]interface{}{
		"ID":          id,
		"IsFavorited": !isFav,
	}); err != nil {
		slog.Error("fav-btn template", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Favorites handles GET /favorites
func Favorites(w http.ResponseWriter, r *http.Request) {
	cards, err := db.ListFavorites()
	if err != nil {
		slog.Error("list favorites", "err", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "favorites.html", map[string]interface{}{
		"Cards": cards,
	}); err != nil {
		slog.Error("favorites template", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Healthz handles GET /healthz
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := db.DB.Ping(); err != nil {
		slog.Error("healthz db ping failed", "err", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"error","detail":%q}`, err.Error())
		return
	}
	fmt.Fprint(w, `{"status":"ok"}`)
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

func pathID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}
