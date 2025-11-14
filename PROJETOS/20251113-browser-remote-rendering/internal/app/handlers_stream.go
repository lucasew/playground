package app

import (
	"net/http"

	"github.com/gorilla/mux"
)

// StreamHTMLHandler handles SSE streaming of HTML content for a specific tab
type StreamHTMLHandler struct {
	App *App
}

func (h *StreamHTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	tabID := vars["tabID"]

	session := h.App.GetOrCreateSession(userID) // Changed to GetOrCreateSession
	if session == nil {
		http.Error(w, "Failed to get or create session", http.StatusInternalServerError)
		return
	}

	// Check if tab exists (this still needs GetTab, as we don't create a tab here)
	_, exists := session.GetTab(tabID)
	if !exists {
		http.Error(w, "Tab not found", http.StatusNotFound)
		return
	}

	h.App.logger.Info("StreamHTMLHandler: Client connected", "userID", userID, "tabID", tabID)
	defer h.App.logger.Info("StreamHTMLHandler: Client disconnected", "userID", userID, "tabID", tabID)

	SetupSSEHeaders(w)

	msgChan := make(chan SSEMessage, 10)
	done := make(chan bool)

	go session.StreamHTML(tabID, msgChan, done)

	for {
		select {
		case msg := <-msgChan:
			if err := SendSSEMessage(w, msg); err != nil {
				h.App.logger.Error("StreamHTMLHandler: Failed to send SSE message", "userID", userID, "tabID", tabID, "error", err)
				close(done)
				return
			}
		case <-r.Context().Done():
			h.App.logger.Info("StreamHTMLHandler: Client context cancelled", "userID", userID, "tabID", tabID)
			close(done)
			return
		}
	}
}

// TabsStreamHandler handles SSE streaming of tab list updates
type TabsStreamHandler struct {
	App *App
}

func (h *TabsStreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	h.App.logger.Debug("TabsStreamHandler: Request received", "vars", vars, "userID", userID)

	session := h.App.GetOrCreateSession(userID) // Changed to GetOrCreateSession
	if session == nil {
		http.Error(w, "Failed to get or create session", http.StatusInternalServerError)
		return
	}

	h.App.logger.Info("TabsStreamHandler: Client connected", "userID", userID)
	defer h.App.logger.Info("TabsStreamHandler: Client disconnected", "userID", userID)

	SetupSSEHeaders(w)

	msgChan := make(chan SSEMessage, 10)

	session.AddListListener(msgChan)
	defer session.RemoveListListener(msgChan)

	tabs := session.GetTabList()
	activeTabID := session.GetActiveTabID()
	initialMsg := SSEMessage{
		Event: "tabs",
		Data: map[string]interface{}{
			"activeTabID": activeTabID,
			"tabs":        tabs,
		},
	}
		if err := SendSSEMessage(w, initialMsg); err != nil {
			h.App.logger.Error("TabsStreamHandler: Failed to send initial SSE message", "userID", userID, "error", err)
			return
		}
		h.App.logger.Debug("TabsStreamHandler: Initial SSE message sent, entering stream loop", "userID", userID)
	
		for {
			select {
			case msg := <-msgChan:
				if err := SendSSEMessage(w, msg); err != nil {
					h.App.logger.Error("TabsStreamHandler: Failed to send SSE message", "userID", userID, "error", err)
					return
				}
			case <-r.Context().Done():
				h.App.logger.Info("TabsStreamHandler: Client context cancelled", "userID", userID)
				return
			}
		}
}
