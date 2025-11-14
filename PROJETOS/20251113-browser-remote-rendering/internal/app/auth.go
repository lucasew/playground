package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json" // New import
	"net/http"

	// New import
	"github.com/gorilla/mux"
)

type contextKey string

const userIDKey contextKey = "userID"

// sessionSecret is a secret key used for generating session tokens.
// In a real application, this should be a strong, randomly generated key
// loaded from a secure configuration.
const sessionSecret = "super-secret-rendereiro-key"

// generateSessionToken creates a simple session token for a given user.
// For an MVP, it's a hash of the username and a secret.
func generateSessionToken(userID string) string {
	hasher := sha256.New()
	hasher.Write([]byte(userID + sessionSecret))
	return hex.EncodeToString(hasher.Sum(nil))
}

// LoginHandler handles user login and sets an authentication cookie
type LoginHandler struct {
	App *App
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !h.App.inputs.Config.ValidateUser(creds.Username, creds.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate session token
	sessionToken := generateSessionToken(creds.Username)

	// Set authentication cookie
	isHTTPS := r.TLS != nil || r.URL.Scheme == "https" || r.Header.Get("X-Forwarded-Proto") == "https"
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,    // Prevent JavaScript access
		Secure:   isHTTPS, // Set Secure based on HTTPS or X-Forwarded-Proto
	}
	http.SetCookie(w, cookie)
	h.App.logger.Debug("LoginHandler: Set-Cookie header sent", "cookie", cookie.String())

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}

// authMiddleware validates HTTP Basic Auth and ensures userID matches authenticated user
func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authenticatedUser string
		var isAuthenticated bool

		// 1. Try Cookie Authentication
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie != nil {
			a.logger.Debug("authMiddleware: Received session_token cookie", "value", cookie.Value) // Log received cookie
			sessionToken := cookie.Value
			// Validate session token by checking against all known users
			for user, _ := range a.inputs.Config.Users {
				if generateSessionToken(user) == sessionToken {
					authenticatedUser = user
					isAuthenticated = true
					break
				}
			}
			if !isAuthenticated {
				a.logger.Debug("authMiddleware: Session token invalid", "sessionToken", sessionToken)
			}
		} else if err != nil && err != http.ErrNoCookie {
			a.logger.Debug("authMiddleware: Error reading session_token cookie", "error", err)
		}

		// 2. If not authenticated by Cookie, try Basic Auth
		if !isAuthenticated {
			user, pass, ok := r.BasicAuth()
			if ok && a.inputs.Config.ValidateUser(user, pass) {
				authenticatedUser = user
				isAuthenticated = true
			}
		}

		if !isAuthenticated {
			w.Header().Set("WWW-Authenticate", `Basic realm="Remote Browser"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check if userID in path matches authenticated user
		vars := mux.Vars(r)
		if pathUser, exists := vars["userID"]; exists {
			a.logger.Debug("authMiddleware: Path UserID", "pathUser", pathUser, "authenticatedUser", authenticatedUser) // Add log
			if pathUser != authenticatedUser {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// Add userID to context
		ctx := context.WithValue(r.Context(), userIDKey, authenticatedUser)
		a.logger.Debug("authMiddleware: Authenticated and passing request to next handler", "userID", authenticatedUser, "path", r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserIDFromContext extracts userID from request context
func getUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
