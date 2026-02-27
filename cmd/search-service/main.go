// search-service exposes a JSON REST API for snake search queries.
// It wraps the iNaturalist API and caches results in-process.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/0xd1sph0l1dus/snakedex/internal/api"
	"github.com/0xd1sph0l1dus/snakedex/internal/middleware"
	"github.com/0xd1sph0l1dus/snakedex/internal/models"
)

type searchResponse struct {
	Cards []models.SnakeCard `json:"cards"`
	Total int                `json:"total"`
	Page  int                `json:"page"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	port := getEnv("PORT", "8081")

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /search", handleSearch)

	slog.Info("search-service starting", "port", port)
	if err := http.ListenAndServe(":"+port, middleware.Metrics(mux)); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, `{"error":"missing q"}`, http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	minObs, _ := strconv.Atoi(r.URL.Query().Get("min_obs"))

	cards, total, err := api.SearchSnakes(q, page, minObs)
	if err != nil {
		slog.Error("search failed", "err", err, "q", q)
		http.Error(w, `{"error":"upstream error"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searchResponse{Cards: cards, Total: total, Page: page})
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
