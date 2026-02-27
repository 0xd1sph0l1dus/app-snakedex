// detail-service exposes a JSON REST API for snake taxon details.
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
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	port := getEnv("PORT", "8082")

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /taxa/{id}", handleDetail)

	slog.Info("detail-service starting", "port", port)
	if err := http.ListenAndServe(":"+port, middleware.Metrics(mux)); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

func handleDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	details, err := api.GetSnakeDetails(id)
	if err != nil {
		slog.Error("detail fetch failed", "err", err, "id", id)
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
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
