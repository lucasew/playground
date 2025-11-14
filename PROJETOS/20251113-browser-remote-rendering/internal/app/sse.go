package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings" // Import strings
	"time"

	"github.com/PuerkitoBio/goquery" // Import goquery
	"github.com/chromedp/chromedp"
	"github.com/microcosm-cc/bluemonday" // Import bluemonday
)

// setupSSEHeaders sets up headers for Server-Sent Events
func setupSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

// SetupSSEHeaders is the public wrapper for setupSSEHeaders
func SetupSSEHeaders(w http.ResponseWriter) {
	setupSSEHeaders(w)
}

// sendSSEMessage sends a single SSE message
func sendSSEMessage(w http.ResponseWriter, msg SSEMessage) error {
	data, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", msg.Event, data)
	if err != nil {
		return err
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// SendSSEMessage is the public wrapper for sendSSEMessage
func SendSSEMessage(w http.ResponseWriter, msg SSEMessage) error {
	return sendSSEMessage(w, msg)
}

// addTabListener adds a listener for tab HTML updates
func (s *UserSession) addTabListener(tabID string, ch chan SSEMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tabListeners[tabID] == nil {
		s.tabListeners[tabID] = make([]chan SSEMessage, 0)
	}
	s.tabListeners[tabID] = append(s.tabListeners[tabID], ch)
}

// removeTabListener removes a listener for tab HTML updates
func (s *UserSession) removeTabListener(tabID string, ch chan SSEMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	listeners := s.tabListeners[tabID]
	for i, listener := range listeners {
		if listener == ch {
			s.tabListeners[tabID] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

// addListListener adds a listener for tab list changes
func (s *UserSession) addListListener(ch chan SSEMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.listListener = append(s.listListener, ch)
}

// removeListListener removes a listener for tab list changes
func (s *UserSession) removeListListener(ch chan SSEMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, listener := range s.listListener {
		if listener == ch {
			s.listListener = append(s.listListener[:i], s.listListener[i+1:]...)
			break
		}
	}
}

// sanitizeAllStyles performs basic sanitization of CSS style values.
func sanitizeAllStyles(css string) string {
	// Remove javascript: from url()
	css = strings.ReplaceAll(css, "url(javascript:", "url(")
	css = strings.ReplaceAll(css, "url( data:image/svg+xml;base64,", "url(data:image/svg+xml;base64,") // Allow SVG data URIs

	// Remove expression()
	css = strings.ReplaceAll(css, "expression(", "")

	return css
}

// unsafeCSS checks if a CSS string contains potentially unsafe content.
func unsafeCSS(css string) bool {
	// Check for javascript: in url()
	if strings.Contains(strings.ToLower(css), "url(javascript:") {
		return true
	}
	// Check for expression()
	if strings.Contains(strings.ToLower(css), "expression(") {
		return true
	}
	return false
}

// streamHTML streams HTML updates for a specific tab
func (s *UserSession) streamHTML(tabID string, msgChan chan SSEMessage, done chan bool) {
	heartbeatTicker := time.NewTicker(10 * time.Second) // Send a heartbeat every 10 seconds
	defer heartbeatTicker.Stop()

	tab, exists := s.getTab(tabID)
	if !exists {
		s.logger.Debug("streamHTML: Tab no longer exists, stopping stream", "userID", s.userID, "tabID", tabID)
		return // Tab no longer exists, stop streaming
	}

	for {
		select {
		case <-tab.domChangeChan: // Listen for DOM change signals
			// Fetch HTML only when a change is signaled
			html, err := tab.getHTML() // This calls chromedp.InnerHTML
			if err != nil {
				if errors.Is(err, context.Canceled) {
					s.logger.Debug("streamHTML: Failed to get HTML for streaming (context canceled)", "userID", s.userID, "tabID", tabID, "error", err)
				} else {
					s.logger.Error("streamHTML: Failed to get HTML for streaming", "userID", s.userID, "tabID", tabID, "error", err)
				}
				continue // Continue trying
			}

			// 1. Sanitiza HTML estrutural mas NÃO deixa o bluemonday tocar no style.
			p := bluemonday.StrictPolicy()
			p.AllowElements("div", "span", "p", "strong", "em", "h1", "h2")
			p.AllowAttrs("class").Globally()
			p.AllowAttrs("id").Globally()
			// NÃO chame AllowStyles() e NÃO permita style aqui.

			intermediate := p.Sanitize(html)

			// 2. Reinjeta styles originais manualmente.
			//    Você mesmo parseia o HTML e copia os style antigos para a versão saneada.

			// O bluemonday removeu os styles. Você precisa extrair antes.
			origDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				s.logger.Error("streamHTML: Failed to parse original HTML with goquery", "userID", s.userID, "tabID", tabID, "error", err)
				continue
			}

			newDoc, err := goquery.NewDocumentFromReader(strings.NewReader(intermediate))
			if err != nil {
				s.logger.Error("streamHTML: Failed to parse intermediate HTML with goquery", "userID", s.userID, "tabID", tabID, "error", err)
				continue
			}

			origStyles := map[string]string{}
			origDoc.Find("*[style]").Each(func(i int, s *goquery.Selection) {
				ptr := fmt.Sprintf("n%d", i)
				s.SetAttr("data-style-id", ptr)
				val, _ := s.Attr("style")
				origStyles[ptr] = val
			})

			// Agora aplique a mesma marcação no doc sanitizado (se a ordem estrutural se manteve).
			newDoc.Find("*").Each(func(i int, s *goquery.Selection) {
				ptr := fmt.Sprintf("n%d", i)
				if style, ok := origStyles[ptr]; ok {
					if !unsafeCSS(style) {
						s.SetAttr("style", style)
					}
				}
			})

			out, err := newDoc.Html()
			if err != nil {
				s.logger.Error("streamHTML: Failed to render final HTML with goquery", "userID", s.userID, "tabID", tabID, "error", err)
				continue
			}
			html = out

			// Compare hash
			hash := sha256Hash(html)
			s.logger.Debug("streamHTML: HTML hash comparison", "userID", s.userID, "tabID", tabID, "currentHash", hash, "lastHash", tab.lastHTMLHash)
			if hash != tab.lastHTMLHash {
				tab.lastHTML = html
				tab.lastHTMLHash = hash

				msg := SSEMessage{
					Event: "html",
					Data:  map[string]string{"content": html},
				}

				select {
				case msgChan <- msg:
					s.logger.Debug("streamHTML: Sent HTML update notification", "userID", s.userID, "tabID", tabID)
				default:
					s.logger.Debug("streamHTML: Skipped sending HTML update notification (channel full or closed)", "userID", s.userID, "tabID", tabID)
				}
			}

		case <-heartbeatTicker.C:
			// Send a heartbeat or just log to keep the connection alive
			s.logger.Debug("streamHTML: Sending heartbeat", "userID", s.userID, "tabID", tabID)
			// Optionally send a "keep-alive" SSE message if the client needs it
			// err := sendSSEMessage(w, SSEMessage{Event: "heartbeat", Data: map[string]string{"time": time.Now().Format(time.RFC3339)}})
			// if err != nil {
			// 	s.logger.Error("streamHTML: Failed to send heartbeat message", "userID", s.userID, "tabID", tabID, "error", err)
			// }

		case <-done:
			s.logger.Debug("streamHTML: Done signal received, stopping stream", "userID", s.userID, "tabID", tabID)
			return
		}
	}
}

// updateTabTitles periodically updates tab titles
func (s *UserSession) updateTabTitles() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			changed := false
			for _, tab := range s.tabs {
				var title string
				err := chromedp.Run(tab.ctx, chromedp.Title(&title))
				if err == nil && title != tab.title {
					tab.title = title
					changed = true
				}
			}
			if changed {
				s.notifyTabListChange()
			}
			s.mu.Unlock()

		case <-s.browserCtx.Done():
			return
		}
	}
}
