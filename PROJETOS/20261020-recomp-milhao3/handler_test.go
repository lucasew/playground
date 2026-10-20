package milhao3

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
			name:     "root is the game page",
			path:     "/",
			status:   http.StatusOK,
			contains: "<title>Show do Milhão 3</title>",
		},
		{
			name:     "stylesheet keeps its query",
			path:     "/css/game.css?v=6",
			status:   http.StatusOK,
			contains: ".stage",
		},
		{
			name:     "script is the file, not the page",
			path:     "/js/game.js?v=6",
			status:   http.StatusOK,
			contains: "const PRIZES",
		},
		{
			name:     "question bank",
			path:     "/js/questions.js",
			status:   http.StatusOK,
			contains: "MILHAO_QUESTIONS",
		},
		{
			name:   "mixed-case image",
			path:   "/assets/img/ajuda1.JPG",
			status: http.StatusOK,
			prefix: []byte{0xff, 0xd8, 0xff},
		},
		{
			name:   "wav from the zip",
			path:   "/assets/audios/Botao.wav",
			status: http.StatusOK,
			prefix: []byte("RIFF"),
		},
		{
			name:     "unknown path falls back to the page",
			path:     "/jogar",
			status:   http.StatusOK,
			contains: "<title>Show do Milhão 3</title>",
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

func TestMediaRange(t *testing.T) {
	handler, err := Handler(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/assets/abertura.mp4", nil)
	request.Header.Set("Range", "bytes=4-7")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
	if got := response.Body.String(); got != "ftyp" {
		t.Fatalf("body = %q, want ftyp", got)
	}
	if got := response.Header().Get("Content-Range"); got != "bytes 4-7/4050309" {
		t.Fatalf("Content-Range = %q", got)
	}
	if got := response.Header().Get("Content-Type"); got != "video/mp4" {
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
