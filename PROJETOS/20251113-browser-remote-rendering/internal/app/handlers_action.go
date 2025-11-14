package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// TabActionHandler handles tab-specific actions
type TabActionHandler struct {
	App *App
}

func (h *TabActionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	tabID := vars["tabID"]

	// Get session
	session, exists := h.App.GetSession(userID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Get tab
	tab, exists := session.GetTab(tabID)
	if !exists {
		http.Error(w, "Tab not found", http.StatusNotFound)
		return
	}

	// Parse action
	var action map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	actionType, ok := action["action"].(string)
	if !ok {
		http.Error(w, "Missing action field", http.StatusBadRequest)
		return
	}

	// Execute action
	var err error
	switch actionType {
	case "navigate":
		url, ok := action["url"].(string)
		if !ok {
			http.Error(w, "Missing url field", http.StatusBadRequest)
			return
		}
		err = tab.Navigate(url)

	case "click":
		x, okX := action["x"].(float64)
		y, okY := action["y"].(float64)
		if !okX || !okY {
			http.Error(w, "Missing x or y coordinates", http.StatusBadRequest)
			return
		}
		err = tab.Click(int(x), int(y))

	case "type":
		text, ok := action["text"].(string)
		if !ok {
			http.Error(w, "Missing text field", http.StatusBadRequest)
			return
		}
		err = tab.TypeText(text)

	case "scrollTo":
		x, okX := action["x"].(float64)
		y, okY := action["y"].(float64)
		if !okX || !okY {
			http.Error(w, "Missing x or y coordinates", http.StatusBadRequest)
			return
		}
		err = tab.ScrollTo(int(x), int(y))

	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// GlobalActionHandler handles global actions (tab management)
type GlobalActionHandler struct {
	App *App
}

func (h *GlobalActionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	// Get or create session
	session := h.App.GetOrCreateSession(userID)
	if session == nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Parse action
	var action map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	actionType, ok := action["action"].(string)
	if !ok {
		http.Error(w, "Missing action field", http.StatusBadRequest)
		return
	}

	// Execute action
	switch actionType {
	case "newTab":
		url, ok := action["url"].(string)
		if !ok {
			http.Error(w, "Missing url field", http.StatusBadRequest)
			return
		}

		tabID, err := session.CreateTab(url)
		if err != nil {
			log.Printf("Error creating new tab for user %s: %v", userID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"tabId":   tabID,
		})

	case "switchTab":
		tabID, ok := action["tabId"].(string)
		if !ok {
			http.Error(w, "Missing tabId field", http.StatusBadRequest)
			return
		}

		err := session.SwitchTab(tabID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

	case "closeTab":
		tabID, ok := action["tabId"].(string)
		if !ok {
			http.Error(w, "Missing tabId field", http.StatusBadRequest)
			return
		}

		err := session.CloseTab(tabID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
	}
}
