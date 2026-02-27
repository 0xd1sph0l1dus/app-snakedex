package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/0xd1sph0l1dus/snakedex/internal/api"
	"github.com/0xd1sph0l1dus/snakedex/internal/client"
	"github.com/0xd1sph0l1dus/snakedex/internal/config"
	"github.com/0xd1sph0l1dus/snakedex/internal/db"
	"github.com/0xd1sph0l1dus/snakedex/internal/handlers"
	"github.com/0xd1sph0l1dus/snakedex/internal/middleware"
	"github.com/0xd1sph0l1dus/snakedex/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	if err := db.Init(cfg.DBPath); err != nil {
		slog.Error("db init failed", "err", err)
		os.Exit(1)
	}

	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		slog.Error("template parse failed", "err", err)
		os.Exit(1)
	}

	// Select backend: microservice clients when env vars are set, direct calls otherwise.
	var s service.Searcher = api.DirectSearcher{}
	var d service.Detailer = api.DirectDetailer{}

	if url := os.Getenv("SEARCH_SERVICE_URL"); url != "" {
		slog.Info("search-service mode", "url", url)
		s = client.NewSearchClient(url)
	} else {
		slog.Info("direct iNaturalist mode (search)")
	}

	if url := os.Getenv("DETAIL_SERVICE_URL"); url != "" {
		slog.Info("detail-service mode", "url", url)
		d = client.NewDetailClient(url)
	} else {
		slog.Info("direct iNaturalist mode (detail)")
	}

	handlers.Init(tmpl, s, d)

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /healthz", handlers.Healthz)
	mux.HandleFunc("GET /", handlers.Index)
	mux.HandleFunc("GET /search", handlers.Search)
	mux.HandleFunc("GET /snake/{id}", handlers.SnakeDetail)
	mux.HandleFunc("POST /snake/{id}/favorite", handlers.ToggleFavorite)
	mux.HandleFunc("GET /favorites", handlers.Favorites)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	slog.Info("snakedex starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, middleware.Metrics(mux)); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
