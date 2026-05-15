// router.go
// All routes

package router

import (
	"database/sql"
	"net/http"

	"github.com/jensma89/mythic-access-gateway/internal/email"
	"github.com/jensma89/mythic-access-gateway/internal/handlers"
	"github.com/jensma89/mythic-access-gateway/internal/middleware"
)

// New builds and returns the main HTTP mux with all routes registered.
func New(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	emailCfg := email.LoadConfig()

	// ---- Rate limiters ----

	// Strict limit for registration endpoints to prevent abuse
	registerLimiter := middleware.NewRateLimiter(0.5, 3) // 1 req / 2s, burst 3

	// Standard limit for all other public routes
	generalLimiter := middleware.NewRateLimiter(10, 20) // 10 req/s, burst 20

	// ---- Public registration routes ----

	mux.Handle("POST /register",
		registerLimiter.Limit(handlers.RegisterHandler(db, emailCfg)),
	)

	// ---- Health check (no auth required) ----

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	// Apply general rate limit to the whole mux as a base layer
	return generalLimiter.Limit(mux)
}
