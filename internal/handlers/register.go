// register.go
// POST /register + GET /verify handler

package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jensma89/mythic-access-gateway/internal/auth"
	"github.com/jensma89/mythic-access-gateway/internal/email"
)

// registerRequest is the JSON body expected from the registration form.
