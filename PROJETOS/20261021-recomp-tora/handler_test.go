package tora

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	handler, err := Handler(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		path     string
		status   int
		contains string
		prefix   []byte
	}{
		{
			name:     "root is the page",
			path:     "/",
			status:   http.StatusOK,
			contains: "<title>TORA</title>",
		},
		{
			name:     "stylesheet keeps its query",
			path:     "/css/app.css?v=1",
			status:   http.StatusOK,
			contains: ".vb-menubar",
		},
		{
			name:     "script is the file, not the page",
			path:     "/js/app.js",
			status:   http.StatusOK,
			contains: "solveLP",
		},
		{
			name:     "rational numbers",
			path:     "/js/rational.js",
			status:   http.StatusOK,
			contains: "export class Rat",
		},
		{
			name:   "sample from the zip",
			path:   "/samples/reddy-mikks.txt",
			status: http.StatusOK,
			prefix: []byte("3000\r\n"),
		},
		{
			name:     "unknown path falls back to the page",
			path:     "/solve",
			status:   http.StatusOK,
			contains: "<title>TORA</title>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			handler.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("GET %s = %d, want %d", tt.path, response.Code, tt.status)
			}
			if got := response.Header().Get("Cache-Control"); got != "max-age=5" {
				t.Fatalf("GET %s Cache-Control = %q", tt.path, got)
			}
			body := response.Body.Bytes()
			if tt.contains != "" && !strings.Contains(string(body), tt.contains) {
				t.Fatalf("GET %s body missing %q", tt.path, tt.contains)
			}
			if len(tt.prefix) > 0 {
				if len(body) < len(tt.prefix) || string(body[:len(tt.prefix)]) != string(tt.prefix) {
					t.Fatalf("GET %s prefix = %x, want %x", tt.path, body[:min(len(body), 8)], tt.prefix)
				}
			}
		})
	}
}

func TestSampleRange(t *testing.T) {
	handler, err := Handler(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/samples/reddy-mikks.txt", nil)
	request.Header.Set("Range", "bytes=0-3")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	if got := response.Body.String(); got != "3000" {
		t.Fatalf("body = %q, want 3000", got)
	}
	if got := response.Header().Get("Content-Range"); got != "bytes 0-3/332" {
		t.Fatalf("Content-Range = %q", got)
	}
	if got := response.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestHandlerPost(t *testing.T) {
	handler, err := Handler(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST / = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}
