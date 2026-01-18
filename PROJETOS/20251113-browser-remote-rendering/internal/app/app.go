package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
	"github.com/lewtec/rendereiro/internal/browser" // Import the new browser package
	"github.com/lewtec/rendereiro/internal/config"
)

// AppInputs contains configuration for the App
type AppInputs struct {
	StateDir    string
	Config      *config.Config
	Port        int
	Host        string
	BrowserPath string // Path to the Chromium executable
	Verbose     bool   // Enable verbose logging (DEBUG level)
}

// App is the main application handler
type App struct {
	inputs   AppInputs
	sessions map[string]*UserSession
	mu       sync.RWMutex
	router   *mux.Router
	logger   *slog.Logger
}

// UserSession represents a user's browser session
type UserSession struct {
	userID      string
	browserCtx  context.Context
	cancelFunc  context.CancelFunc
	profileDir  string
	tabs        map[string]*Tab
	activeTabID string
	mu          sync.RWMutex
	logger      *slog.Logger

	// SSE clients
	tabListeners map[string][]chan SSEMessage // tabID -> list of SSE channels
	listListener []chan SSEMessage            // listeners for tab list changes
}

// Tab represents a browser tab
type Tab struct {
	id            string
	ctx           context.Context
	cancelFunc    context.CancelFunc
	url           string
	title         string
	lastHTML      string
	lastHTMLHash  string
	domChangeChan chan struct{} // Channel to signal DOM changes
}

// SSEMessage represents a Server-Sent Event message
type SSEMessage struct {
	Event string      // "html" or "tabs"
	Data  interface{} // will be serialized as JSON
}

// New creates a new App instance
func New(inputs AppInputs) (*App, error) {
	// Create state directory if it doesn't exist
	if err := os.MkdirAll(inputs.StateDir, 0755); err != nil {
		return nil, fmt.Errorf("create state directory: %w", err)
	}

	logLevel := slog.LevelInfo
	if inputs.Verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	app := &App{
		inputs:   inputs,
		sessions: make(map[string]*UserSession),
		router:   mux.NewRouter(),
		logger:   logger,
	}

	app.setupRoutes()
	return app, nil
}

// CheckSessionHandler simply returns 200 OK if the request is authenticated
type CheckSessionHandler struct {
	App *App
}

func (h *CheckSessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If we reach here, authMiddleware has already authenticated the request
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		// This should ideally not happen if authMiddleware worked correctly
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"userID": userID})
}

// setupRoutes configures all HTTP routes
func (a *App) setupRoutes() {
	// Public routes
	a.router.HandleFunc("/api/health", HealthHandler).Methods("GET")
	a.router.Handle("/api/login", &LoginHandler{App: a}).Methods("POST")

	// Protected routes (with auth middleware)
	protected := a.router.PathPrefix("/api").Subrouter()
	protected.Use(a.authMiddleware)

	// New protected route for session check
	protected.Handle("/check-session", &CheckSessionHandler{App: a}).Methods("GET")

	// SSE streams
	protected.Handle("/stream/{userID}/{tabID}", &StreamHTMLHandler{App: a}).Methods("GET")
	protected.Handle("/tabs/{userID}", &TabsStreamHandler{App: a}).Methods("GET")

	// Actions
	protected.Handle("/action/{userID}/{tabID}", &TabActionHandler{App: a}).Methods("POST")
	protected.Handle("/action/{userID}", &GlobalActionHandler{App: a}).Methods("POST")

	// Static files (must be last)
	a.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))
}

// ServeHTTP implements http.Handler
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

// getOrCreateSession retrieves or creates a user session
func (a *App) getOrCreateSession(userID string) *UserSession {
	a.mu.Lock()
	defer a.mu.Unlock()

	if session, exists := a.sessions[userID]; exists {
		return session
	}

	profileDir := filepath.Join(a.inputs.StateDir, "chrome-profile-"+userID)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		a.logger.Error("failed to create profile directory", "userID", userID, "error", err)
		return nil
	}

	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.UserDataDir(profileDir),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	}

	var browserPath string // Declare browserPath locally
	if a.inputs.BrowserPath != "" {
		browserPath = a.inputs.BrowserPath
	} else {
		browserPath = browser.FindChromiumExecutable()
	}

	if browserPath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(browserPath))
		a.logger.Info("Using Chromium executable", "path", browserPath) // Log the path
	} else {
		a.logger.Warn("Chromium executable not found. chromedp will attempt to use its default discovery which might fail.")
	}

	// Add verbose chromedp logging if enabled
	// chromedp.WithDebugf is for ContextOption, not ExecAllocatorOption.
	// We will apply chromedp.WithLogf to NewContext instead.

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	if allocCtx == nil { // Check if allocCtx is nil, indicating an error
		a.logger.Error("Failed to create new exec allocator", "userID", userID)
		return nil
	}

	var browserCtxOpts []chromedp.ContextOption
	if a.inputs.Verbose {
		browserCtxOpts = append(browserCtxOpts, chromedp.WithLogf(a.logger.Debug))
	}
	browserCtx, _ := chromedp.NewContext(allocCtx, browserCtxOpts...)

	session := &UserSession{
		userID:       userID,
		browserCtx:   browserCtx,
		cancelFunc:   cancel,
		profileDir:   profileDir,
		tabs:         make(map[string]*Tab),
		tabListeners: make(map[string][]chan SSEMessage),
		listListener: make([]chan SSEMessage, 0),
		logger:       a.logger,
	}

	a.sessions[userID] = session
	a.logger.Info("created new user session", "userID", userID)

	return session
}

// GetSession retrieves an existing user session
func (a *App) GetSession(userID string) (*UserSession, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	session, exists := a.sessions[userID]
	return session, exists
}

// Cleanup performs cleanup of resources
func (a *App) Cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for userID, session := range a.sessions {
		session.cleanup()
		a.logger.Info("cleaned up session", "userID", userID)
	}
}

// cleanup performs cleanup of a user session
func (s *UserSession) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close all tabs
	for _, tab := range s.tabs {
		tab.cancelFunc()
	}

	// Cancel browser context
	s.cancelFunc()

	// Close all SSE channels
	for _, listeners := range s.tabListeners {
		for _, ch := range listeners {
			close(ch)
		}
	}
	for _, ch := range s.listListener {
		close(ch)
	}
}

// Public methods for handlers

// GetOrCreateSession retrieves or creates a user session (public)
func (a *App) GetOrCreateSession(userID string) *UserSession {
	return a.getOrCreateSession(userID)
}

// GetTab retrieves a tab by ID (public wrapper)
func (s *UserSession) GetTab(tabID string) (*Tab, bool) {
	return s.getTab(tabID)
}

// GetTabList returns a list of all tabs (public wrapper)
func (s *UserSession) GetTabList() []map[string]interface{} {
	return s.getTabList()
}

// GetActiveTabID returns the active tab ID
func (s *UserSession) GetActiveTabID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeTabID
}

// CreateTab creates a new tab (public wrapper)
func (s *UserSession) CreateTab(url string) (string, error) {
	return s.createTab(url)
}

// CloseTab closes a tab (public wrapper)
func (s *UserSession) CloseTab(tabID string) error {
	return s.closeTab(tabID)
}

// SwitchTab switches active tab (public wrapper)
func (s *UserSession) SwitchTab(tabID string) error {
	return s.switchTab(tabID)
}

// AddListListener adds a listener for tab list changes (public wrapper)
func (s *UserSession) AddListListener(ch chan SSEMessage) {
	s.addListListener(ch)
}

// RemoveListListener removes a listener for tab list changes (public wrapper)
func (s *UserSession) RemoveListListener(ch chan SSEMessage) {
	s.removeListListener(ch)
}

// StreamHTML streams HTML updates (public wrapper)
func (s *UserSession) StreamHTML(tabID string, msgChan chan SSEMessage, done chan bool) {
	s.streamHTML(tabID, msgChan, done)
}
