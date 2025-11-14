package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"            // Import os
	"path/filepath" // Import path/filepath
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

// createTab creates a new browser tab and navigates to the given URL
func (s *UserSession) createTab(url string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error // Declare err here

	s.logger.Info("Attempting to create new tab", "userID", s.userID, "url", url) // Entry log

	tabID := uuid.New().String()
	tabCtx, cancel := chromedp.NewContext(s.browserCtx)

	tab := &Tab{
		id:            tabID,
		ctx:           tabCtx,
		cancelFunc:    cancel,
		url:           url,
		domChangeChan: make(chan struct{}, 1), // Initialize the channel
	}

	// Enable DOM event listening
	err = chromedp.Run(tabCtx, dom.Enable())
	if err != nil {
		s.logger.Error("Failed to enable DOM event listening", "userID", s.userID, "tabID", tabID, "error", err)
		cancel()
		return "", err
	}

	// Start goroutine to listen for DOM changes
	go func() {
		s.logger.Debug("Starting DOM event listener for tab", "userID", s.userID, "tabID", tabID)
		chromedp.ListenTarget(tab.ctx, func(ev interface{}) {
			switch ev := ev.(type) {
			case *dom.EventSetChildNodes:
				s.logger.Debug("DOM.EventSetChildNodes received", "userID", s.userID, "tabID", tabID, "parentID", ev.ParentID)
				select {
				case tab.domChangeChan <- struct{}{}:
				default:
					// Channel is full, a signal is already pending, no need to send another
				}
			case *dom.EventChildNodeInserted:
				s.logger.Debug("DOM.EventChildNodeInserted received", "userID", s.userID, "tabID", tabID, "parentNodeID", ev.ParentNodeID, "nodeName", ev.Node.NodeName)
				select {
				case tab.domChangeChan <- struct{}{}:
				default:
					// Channel is full, a signal is already pending, no need to send another
				}
			case *dom.EventChildNodeRemoved:
				s.logger.Debug("DOM.EventChildNodeRemoved received", "userID", s.userID, "tabID", tabID, "parentNodeID", ev.ParentNodeID, "nodeID", ev.NodeID)
				select {
				case tab.domChangeChan <- struct{}{}:
				default:
					// Channel is full, a signal is already pending, no need to send another
				}
			}
		})
		s.logger.Debug("Stopping DOM event listener for tab (context done)", "userID", s.userID, "tabID", tabID)
	}()

	// Navigate and capture title with timeout
	ctx, cancelTimeout := context.WithTimeout(tabCtx, 30*time.Second) // Increased timeout to 30 seconds
	defer cancelTimeout()

	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Title(&tab.title),
	)

	if err != nil {
		s.logger.Error("Failed to create new tab during navigation or title capture", "userID", s.userID, "url", url, "error", err) // Error log
		cancel()
		return "", err
	}

	// Capture screenshot for debugging
	var buf []byte
	if err := chromedp.Run(tabCtx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		s.logger.Error("Failed to capture screenshot", "userID", s.userID, "tabID", tabID, "error", err)
	} else {
		screenshotPath := filepath.Join(s.profileDir, fmt.Sprintf("screenshot-%s.png", tabID))
		if err := os.WriteFile(screenshotPath, buf, 0644); err != nil {
			s.logger.Error("Failed to save screenshot", "userID", s.userID, "tabID", tabID, "error", err)
		} else {
			s.logger.Info("Screenshot captured", "userID", s.userID, "tabID", tabID, "path", screenshotPath)
		}
	}

	s.tabs[tabID] = tab
	if s.activeTabID == "" {
		s.activeTabID = tabID
	}

	s.notifyTabListChange()
	s.logger.Info("Successfully created new tab", "userID", s.userID, "tabID", tabID, "url", url) // Success log

	// Send an initial signal to trigger HTML stream
	select {
	case tab.domChangeChan <- struct{}{}:
	default:
	}

	return tabID, nil
}

// closeTab closes a browser tab
func (s *UserSession) closeTab(tabID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tab, exists := s.tabs[tabID]
	if !exists {
		return errors.New("tab not found")
	}

	tab.cancelFunc()
	delete(s.tabs, tabID)

	// If this was the active tab, select another one
	if s.activeTabID == tabID {
		s.activeTabID = ""
		for id := range s.tabs {
			s.activeTabID = id
			break
		}
	}

	s.notifyTabListChange()
	return nil
}

// switchTab changes the active tab
func (s *UserSession) switchTab(tabID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tabs[tabID]; !exists {
		return errors.New("tab not found")
	}

	s.activeTabID = tabID
	s.notifyTabListChange()
	return nil
}

// getTab retrieves a tab by ID
func (s *UserSession) getTab(tabID string) (*Tab, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tab, exists := s.tabs[tabID]
	return tab, exists
}

// getTabList returns a list of all tabs
func (s *UserSession) getTabList() []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tabs := make([]map[string]interface{}, 0, len(s.tabs))
	for _, tab := range s.tabs {
		tabs = append(tabs, map[string]interface{}{
			"id":    tab.id,
			"title": tab.title,
			"url":   tab.url,
		})
	}

	return tabs
}

// notifyTabListChange sends notification to all tab list listeners
func (s *UserSession) notifyTabListChange() {
	// This will be called with mutex already locked, so we don't lock again
	msg := SSEMessage{
		Event: "tabs",
		Data: map[string]interface{}{
			"activeTabID": s.activeTabID,
			"tabs":        s.getTabsForNotification(),
		},
	}

	for _, ch := range s.listListener {
		select {
		case ch <- msg:
			s.logger.Debug("Sent tab list change notification", "userID", s.userID)
		default:
			s.logger.Debug("Skipped sending tab list change notification (channel full or closed)", "userID", s.userID)
		}
	}
}

// getTabsForNotification returns tabs data for notification (called with lock held)
func (s *UserSession) getTabsForNotification() []map[string]interface{} {
	tabs := make([]map[string]interface{}, 0, len(s.tabs))
	for _, tab := range s.tabs {
		tabs = append(tabs, map[string]interface{}{
			"id":    tab.id,
			"title": tab.title,
			"url":   tab.url,
		})
	}
	return tabs
}

// navigate navigates a tab to a new URL
func (t *Tab) navigate(url string) error {
	ctx, cancel := context.WithTimeout(t.ctx, 10*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
	)

	if err != nil {
		return err
	}

	t.url = url
	// Update title
	chromedp.Run(ctx, chromedp.Title(&t.title))

	// Signal DOM change after navigation
	select {
	case t.domChangeChan <- struct{}{}:
	default:
	}

	return nil
}

// click performs a click at the given coordinates
func (t *Tab) click(x, y int) error {
	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.MouseClickXY(float64(x), float64(y)))
	if err != nil {
		return err
	}

	// Signal DOM change after click
	select {
	case t.domChangeChan <- struct{}{}:
	default:
	}

	return nil
}

// typeText types text into the currently focused element
func (t *Tab) typeText(text string) error {
	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.SendKeys(":focus", text, chromedp.ByQuery))
	if err != nil {
		return err
	}

	// Signal DOM change after typing
	select {
	case t.domChangeChan <- struct{}{}:
	default:
	}

	return nil
}

// scrollTo scrolls to absolute position
func (t *Tab) scrollTo(x, y int) error {
	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	script := fmt.Sprintf("window.scrollTo(%d, %d)", x, y)
	err := chromedp.Run(ctx, chromedp.Evaluate(script, nil))
	if err != nil {
		return err
	}

	// Signal DOM change after scroll
	select {
	case t.domChangeChan <- struct{}{}:
	default:
	}

	return nil
}

// getHTML retrieves the current HTML content
func (t *Tab) getHTML() (string, error) {
	ctx, cancel := context.WithTimeout(t.ctx, 2*time.Second)
	defer cancel()

	var html string
	err := chromedp.Run(ctx, chromedp.InnerHTML("html", &html, chromedp.ByQuery))
	if err != nil {
		return "", err
	}

	return html, nil
}

// sha256Hash computes SHA256 hash of a string
func sha256Hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// Public methods for handlers

// Navigate navigates a tab to a new URL (public wrapper)
func (t *Tab) Navigate(url string) error {
	return t.navigate(url)
}

// Click performs a click at the given coordinates (public wrapper)
func (t *Tab) Click(x, y int) error {
	return t.click(x, y)
}

// TypeText types text into the currently focused element (public wrapper)
func (t *Tab) TypeText(text string) error {
	return t.typeText(text)
}

// ScrollTo scrolls to absolute position (public wrapper)
func (t *Tab) ScrollTo(x, y int) error {
	return t.scrollTo(x, y)
}
