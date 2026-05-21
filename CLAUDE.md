# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

FES (Festival Edit System) is a Japanese CMS for managing festival/event websites. It consists of a Go backend deployed on Google App Engine and a React maps SPA.

## Build & Run Commands

### Backend (Go)

```powershell
# Run locally (from app/ directory)
cd app
go run ./_cmd/main.go

# Build binary
go build -o app ./_cmd

# Deploy to Google App Engine
gcloud app deploy app/app.yaml
```

### Frontend (React/Vite — maps/)

```powershell
cd maps
npm install
npm run dev       # Dev server (base path: /maps/)
npm run build     # Output to maps/dist/
npm run lint      # ESLint (0 warnings allowed)
npm run preview   # Preview production build
```

### Static Site Archive

```go
// logic.CreateStaticSite() generates a full static HTML snapshot
// The _cmd/archive/ directory contains the headless archive tool
```

## Architecture

### Two-part application

1. **`app/`** — Go backend (App Engine, Go 1.24)
2. **`maps/`** — React SPA served at the `/maps/` path from the built `dist/`

### Backend layers

| Package | Role |
|---|---|
| `app/handler/` | HTTP handlers for public routes (pages, files, auth) |
| `app/handler/manage/` | Admin console routes |
| `app/handler/internal/` | Cache server, maps integration, static asset serving, archive |
| `app/datastore/` | DAO layer — one file per Datastore kind |
| `app/logic/` | Business logic (static site generation, HTML rendering) |
| `app/api/` | Template helpers (`html/template` FuncMap, content conversion) |
| `app/config/` | Functional-option pattern for app configuration |

Routes are registered in `handler.Register()` (public) and `manage.Register()` (admin) via Gorilla Mux.

### Data model (Google Cloud Datastore)

Core kinds: `Page` / `PageData`, `HTML` (cached render), `File` / `FileData`, `Template`, `Variable`, `Site`, `Draft`, `Meta`.

Pages form a parent-child tree. `HTML` stores pre-rendered output to avoid re-rendering on each request.

### Embedded assets

`app/handler/internal/_assets/` is embedded with `//go:embed`. It contains:
- `environment.json` — OAuth2 credentials (CLIENT_ID, CLIENT_SECRET)
- `maps/` — built React app files copied from `maps/dist/`
- `archives/` — static archive zips

After running `npm run build` in `maps/`, copy the `dist/` output into `app/handler/internal/_assets/maps/` before building or deploying the Go binary.

### Maps frontend (maps/src/)

- `App.jsx` — root component, layout
- `Map.jsx` — HTML5 Canvas map renderer with responsive scaling
- `MapPoint.jsx` — shop geolocation data and point definitions
- `SVGButton.jsx` — reusable SVG icon button
- `Window.jsx` — window resize hook

Vite is configured with `base: '/maps/'` so all asset paths are rooted there.

### Authentication

Google OAuth2 via credentials embedded in `environment.json`. Sessions managed with Gorilla sessions. JWT used for token handling.

## Key Conventions

- Management UI, comments, and commit messages are in Japanese.
- The DAO layer separates metadata entities from data entities (e.g., `Page` + `PageData`, `File` + `FileData`) to reduce read costs.
- Template rendering uses Go's `html/template` with custom functions registered in `api/helper.go`.
- No test files exist in the repo; manual/integration testing is the current approach.
