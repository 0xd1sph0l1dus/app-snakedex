package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/0xd1sph0l1dus/snakedex/internal/handlers"
)

func main() {
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}
	handlers.Init(tmpl)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handlers.Index)
	mux.HandleFunc("GET /search", handlers.Search)
	mux.HandleFunc("GET /snake/{id}", handlers.SnakeDetail)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Snakedex listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}
