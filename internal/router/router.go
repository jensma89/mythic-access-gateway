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
