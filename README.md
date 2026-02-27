# 🐍 Snakedex

Encyclopédie web des serpents du monde, construite en Go avec HTMX.
Données fournies par l'[API iNaturalist](https://api.inaturalist.org/v1/docs/).

---

## Stack

| Couche | Technologie |
|---|---|
| Backend | Go 1.24 — stdlib + `modernc.org/sqlite` + `prometheus/client_golang` |
| Frontend | HTMX 1.9 · HTML templates · CSS custom |
| Cartographie | Leaflet.js (CDN) |
| Base de données | SQLite (favoris, historique) |
| Cache | `sync.Map` in-process (TTL 5 min search / 30 min detail) |
| Observabilité | `log/slog` JSON · Prometheus `/metrics` |
| Infra | Docker multi-stage · docker-compose · Nginx · GitHub Actions |

---

## Lancer le projet

### Développement (sans Docker)

```bash
cd snakedex
go run ./cmd/server
# → http://localhost:8080
```

Variables d'environnement disponibles :

| Variable | Défaut | Description |
|---|---|---|
| `PORT` | `8080` | Port d'écoute |
| `DB_PATH` | `snakedex.db` | Chemin du fichier SQLite |
| `SEARCH_SERVICE_URL` | _(vide)_ | URL du search-service (mode microservices) |
| `DETAIL_SERVICE_URL` | _(vide)_ | URL du detail-service (mode microservices) |

### Production (Docker Compose)

```bash
docker compose up --build
```

| Service | URL publique | Description |
|---|---|---|
| App web | http://localhost | Via Nginx (port 80) |
| Prometheus | http://localhost:9090 | Métriques |
| search-service | interne `:8081` | JSON API recherche |
| detail-service | interne `:8082` | JSON API détail |

---

## Architecture

```
browser
  └── Nginx :80
        └── frontend :8080   (HTML + SQLite)
              ├── search-service :8081   (GET /search → JSON)
              └── detail-service :8082   (GET /taxa/{id} → JSON)

Prometheus :9090  ←  scrape /metrics  (frontend, search-service, detail-service)
```

### Structure des dossiers

```
snakedex/
├── cmd/
│   ├── server/          # Frontend web (monolithe ou orchestrateur)
│   ├── search-service/  # JSON API — recherche iNaturalist
│   └── detail-service/  # JSON API — détail taxon iNaturalist
├── internal/
│   ├── api/             # Client HTTP iNaturalist + cache sync.Map
│   ├── cache/           # Cache TTL générique (sync.Map)
│   ├── client/          # Clients HTTP inter-services (SearchClient, DetailClient)
│   ├── config/          # Config par variables d'environnement
│   ├── db/              # SQLite — favoris + historique de recherche
│   ├── handlers/        # Handlers HTTP (Index, Search, Detail, Favorites…)
│   ├── middleware/       # Prometheus metrics middleware
│   ├── models/          # SnakeCard, SnakeDetails, Coordinate
│   └── service/         # Interfaces Searcher / Detailer
├── templates/           # HTML templates HTMX
├── static/              # CSS, JS (map.js)
├── Dockerfile           # Image frontend (multi-stage)
├── Dockerfile.service   # Image générique services (ARG CMD)
├── docker-compose.yml
├── nginx.conf
└── prometheus.yml
```

### Mode monolithe vs microservices

Le frontend sélectionne son backend au démarrage :

- **Sans** `SEARCH_SERVICE_URL` / `DETAIL_SERVICE_URL` → appels directs à iNaturalist (`api.DirectSearcher`)
- **Avec** les env vars → délègue aux services via HTTP (`client.SearchClient`)

---

## API interne

### search-service (`:8081`)

```
GET /search?q=python&page=1&min_obs=100
→ { "cards": [...], "total": 42, "page": 1 }

GET /healthz
→ { "status": "ok" }

GET /metrics
→ Prometheus text format
```

### detail-service (`:8082`)

```
GET /taxa/{id}
→ SnakeDetails JSON

GET /healthz → { "status": "ok" }
GET /metrics → Prometheus text format
```

### Nginx (routes publiques)

```
GET /            → frontend
GET /api/search  → search-service
GET /api/taxa/*  → detail-service
GET /metrics     → 403 (bloqué)
```

---

## Fonctionnalités

- **Recherche** — texte libre, filtre min. observations, tri
- **Pagination** — 20 résultats par page
- **Détail** — taxonomie, résumé Wikipedia, galerie photos, carte Leaflet
- **Favoris** — sauvegarde persistante en SQLite, page dédiée
- **Historique** — chaque recherche est loggée en base
- **Cache** — réponses iNaturalist mises en cache (5-30 min)
- **Métriques** — `http_requests_total` + `http_request_duration_seconds` par service

---

## Roadmap

- [x] Phase 1 — Fonctionnalités (pagination, filtres, détail enrichi)
- [x] Phase 2 — Persistence (SQLite, cache sync.Map)
- [x] Phase 3 — DevOps (Docker, CI/CD, healthz, env config)
- [x] Phase 4 — Microservices (slog, Prometheus, services indépendants, Nginx)
- [ ] gRPC (remplacement du transport HTTP inter-services)
- [ ] Tracing distribué (OpenTelemetry)
- [ ] Redis (cache distribué partagé entre services)
