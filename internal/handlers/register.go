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
type registerRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// RegisterHandler handles POST /register.
// It creates an unverified user and sends a verification email.
func RegisterHandler(db *sql.DB, emailCfg email.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST
		if r.Method != http.MethodPost {
			jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Basic input validation
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		req.Name = strings.TrimSpace(req.Name)

		if req.Email == "" || !strings.Contains(req.Email, "@") {
			jsonError(w, "valid email required", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			jsonError(w, "name required", http.StatusBadRequest)
			return
		}

		// Check if email is already registered
		var exists int
		err := db.QueryRowContext(r.Context(), `SELECT COUNT(*) 
		FROM external_users WHERE email = ?`, req.Email).Scan(&exists)
		if err != nil {
			jsonError(w, "internal error", http.StatusInternalServerError)
			return
		}

		if exists > 0 {
			// Return 200 to avoid leaking which emails are registered
			jsonOK(w, "if this email is not yet registered you will receive a verification link")
			return
		}

		// Generate a one-time verification token valid for 24 hours
		token, err := auth.GenerateVerificationToken()
		if err != nil {
			jsonError(w, "internal error", http.StatusInternalServerError)
			return
		}

		expires := time.Now().Add(24 * time.Hour).Unix()

		// Insert the unverified user
		_, err = db.ExecContext(r.Context(), `
INSERT INTO external_users (email, name, verification_token, verification_expires)
VALUES (?, ?, ?, ?)`,
			req.Email, req.Name, token, expires,
		)
		if err != nil {
			jsonError(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Send verification email (non-blocking failure - log but don't expose)
		if err := email.SendVerification(emailCfg, req.Email, req.Name, token); err != nil {

			// Log the error server-side but keep the response generic
			_ = err // TODO: replace with proper logger when available
			jsonError(w, "failed to send verification email", http.StatusInternalServerError)
			return
		}

		jsonOK(w, "verification email sent - check your inbox")
	}
}

// VerifyHandler handles GET /verify?token=<token>.
// On success it creates the shadow account in FastAPI and emails the API key.
func VerifyHandler(db *sql.DB, emailCfg email.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			jsonError(w, "token required", http.StatusBadRequest)
			return
		}

		// Look up the user by token and check expiry
		var (
			userID   int64
			email_   string
			name     string
			expires  int64
			verified int
		)
		err := db.QueryRowContext(r.Context(), `
SELECT id, email, name, verification_expires, verified
FROM external_users
WHERE verification_token = ?`,
			token).Scan(&userID, &email_, &name, &expires, &verified)

		if err == sql.ErrNoRows {
			jsonError(w, "invalid or expired token", http.StatusBadRequest)
			return
		}

		if err != nil {
			jsonError(w, "internal error", http.StatusInternalServerError)
			return
		}

		if verified == 1 {
			jsonError(w, "email already verified", http.StatusBadRequest)
			return
		}

		if time.Now().Unix() > expires {
			jsonError(w, "token expired - please register again", http.StatusBadRequest)
			return
		}

		// Generate the raw API key - shown to user once via email
		apiKey, err := auth.GenerateAPIKey()
		if err != nil {
			jsonError(w, "internal error", http.StatusInternalServerError)
			return
		}

		keyHash, err := auth.HashAPIKey(apiKey)
		if err != nil {
			jsonError(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Mark user as verified, store the key hash, clear the token
		_, err = db.ExecContext(r.Context(), `
UPDATE external_users
SET verified = 1, 
    api_key_hash = ?, 
    verification_token ='',
    verification_expires = 0,
WHERE id = ?`, keyHash, userID,
		)
		if err != nil {
			jsonError(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Send the API key by email
		if err := email.SendAPIKey(emailCfg, email_, name, apiKey); err != nil {
			jsonError(2, "failed to send API key email", http.StatusInternalServerError)
			return
		}

		jsonOK(w, "email verified - your API key has been sent to your inbox")
	}
}

// ---- helpers ----
