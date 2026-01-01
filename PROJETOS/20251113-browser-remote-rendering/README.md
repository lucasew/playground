# Rendereiro - Remote Browser Service

Experimento de site que mastiga outros sites com um navegador headless.

A Go-based service that spawns isolated headless browser instances per user and streams HTML to remote clients via Server-Sent Events (SSE). Each user has an isolated profile with independently managed tabs.

## Features

- **Multi-user Support**: Isolated browser sessions per user with persistent profiles
- **Multiple Tabs**: Each user can manage multiple browser tabs independently
- **Real-time Streaming**: HTML content streamed via Server-Sent Events (SSE)
- **Tab Management**: Create, close, and switch between tabs
- **Browser Controls**: Navigate, click, type, and scroll in remote browser
- **HTTP Basic Auth**: Simple authentication system
- **PWA Client**: Progressive Web App interface for desktop and mobile

## Architecture

### Stack

- **Backend**: Go + chromedp + cobra + gorilla/mux
- **Frontend**: HTML/CSS/JS (PWA)
- **Auth**: HTTP Basic Auth
- **Config**: YAML
- **Streaming**: Server-Sent Events (SSE)

### Components

- **CLI**: Cobra-based command-line interface
- **Browser Manager**: chromedp for headless Chrome control
- **Session Manager**: Isolated user sessions with persistent profiles
- **Tab Manager**: Multiple tabs per user session
- **SSE Streaming**: Real-time HTML updates and tab list synchronization
- **Actions API**: RESTful API for browser interactions

## Quick Start

### Prerequisites

- Go 1.21 or later
- Chrome/Chromium installed (for local development)

### Installation

```bash
# Clone the repository
git clone https://github.com/lewtec/rendereiro.git
cd rendereiro

# Install dependencies
go mod tidy

# Build
go build -o rendereiro cmd/rendereiro/main.go

# Run
./rendereiro serve
```

The server will start on `http://localhost:8080`

### Configuration

Create a `config.yaml` file:

```yaml
users:
  alice: senha123
  bob: outrasenha
  carlos: maisuma
```

**Note**: Passwords are stored in plaintext in the MVP. Use bcrypt for production.

### Command-Line Options

```bash
rendereiro serve [flags]

Flags:
  --config string   Path to config.yaml (default "config.yaml")
  --state string    Directory for Chrome profiles (default "./profiles")
  --port int        HTTP server port (default 8080)
  --host string     Host to bind to (default "0.0.0.0")
```

**Examples:**

```bash
# Run with default settings
rendereiro serve

# Custom configuration
rendereiro serve --config /etc/rendereiro/config.yaml --port 3000

# Custom state directory
rendereiro serve --state /var/lib/rendereiro/profiles
```

## API Reference

### Authentication

All API endpoints (except `/` and `/api/health`) require HTTP Basic Auth.

```
Authorization: Basic base64(username:password)
```

The `userID` in the URL path must match the authenticated username.

### Endpoints

#### Health Check

```
GET /api/health
```

Returns server health status.

**Response:**
```json
{"status": "ok"}
```

---

#### HTML Stream (SSE)

```
GET /api/stream/{userID}/{tabID}
```

Streams HTML updates for a specific tab.

**Headers:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Event:**
```
event: html
data: {"content": "<html>...</html>"}
```

---

#### Tabs Stream (SSE)

```
GET /api/tabs/{userID}
```

Streams tab list updates.

**Event:**
```
event: tabs
data: {
  "activeTabID": "tab-123",
  "tabs": [
    {"id": "tab-123", "title": "Google", "url": "https://google.com"},
    {"id": "tab-456", "title": "GitHub", "url": "https://github.com"}
  ]
}
```

---

#### Tab Actions

```
POST /api/action/{userID}/{tabID}
```

Execute actions on a specific tab.

**Navigate:**
```json
{"action": "navigate", "url": "https://example.com"}
```

**Click:**
```json
{"action": "click", "x": 100, "y": 200}
```

**Type:**
```json
{"action": "type", "text": "hello world"}
```

**Scroll:**
```json
{"action": "scrollTo", "x": 0, "y": 500}
```

---

#### Global Actions

```
POST /api/action/{userID}
```

Manage tabs globally.

**New Tab:**
```json
{"action": "newTab", "url": "https://google.com"}
```

**Response:**
```json
{"success": true, "tabId": "tab-789"}
```

**Switch Tab:**
```json
{"action": "switchTab", "tabId": "tab-123"}
```

**Close Tab:**
```json
{"action": "closeTab", "tabId": "tab-456"}
```

## Docker Deployment

### Build Image

```bash
docker build -t rendereiro .
```

### Run Container

```bash
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/profiles:/app/profiles \
  rendereiro
```

### Docker Compose

```yaml
version: '3.8'

services:
  rendereiro:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml
      - ./profiles:/app/profiles
    environment:
      - TZ=America/Sao_Paulo
```

## Usage

1. **Access the Web Interface**: Navigate to `http://localhost:8080`

2. **Login**: Enter your username and password (from `config.yaml`)

3. **Create a Tab**: Click the "+" button and enter a URL

4. **Browse**: The remote browser will navigate to the URL and stream the content

5. **Interact**: Click, type, and scroll in the streamed content

6. **Manage Tabs**: Switch between tabs or close them using the tab bar

## Development

### Project Structure

```
rendereiro/
├── cmd/
│   └── rendereiro/
│       └── main.go                 # Entry point + CLI
├── internal/
│   ├── app/
│   │   ├── app.go                 # App struct + routing
│   │   ├── auth.go                # Auth middleware
│   │   ├── tab.go                 # Tab management
│   │   ├── sse.go                 # SSE helpers
│   │   ├── handlers_stream.go     # SSE handlers
│   │   ├── handlers_action.go     # Action handlers
│   │   └── handlers_health.go     # Health check
│   └── config/
│       └── config.go              # Config loading
├── web/
│   ├── index.html                 # Client UI
│   ├── app.js                     # Client logic
│   ├── style.css                  # Styling
│   └── manifest.json              # PWA manifest
├── config.yaml                    # User configuration
├── go.mod
├── go.sum
├── Dockerfile
└── README.md
```

### Running Tests

```bash
go test ./...
```

## Security Considerations

**Current MVP limitations:**

- Passwords stored in plaintext in `config.yaml`
- No TLS/HTTPS enforcement
- Basic HTTP authentication only
- No rate limiting

**Production recommendations:**

- Use bcrypt for password hashing
- Enable HTTPS/TLS
- Implement JWT or OAuth2
- Add rate limiting per user
- Configure CORS properly
- Use Content Security Policy (CSP)

## Future Improvements

### Security
- [ ] Bcrypt password hashing
- [ ] JWT authentication
- [ ] TLS/HTTPS support
- [ ] Rate limiting
- [ ] CORS configuration

### Performance
- [ ] DOM diffing instead of full HTML
- [ ] HTML compression (gzip)
- [ ] Event debouncing
- [ ] Browser instance pooling

### Features
- [ ] File upload/download
- [ ] Screenshots on-demand
- [ ] Session persistence
- [ ] Browser history
- [ ] Shared bookmarks
- [ ] Multi-user collaboration

### Observability
- [ ] Prometheus metrics
- [ ] Structured logging levels
- [ ] Distributed tracing
- [ ] Admin dashboard

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Acknowledgments

Built with:
- [chromedp](https://github.com/chromedp/chromedp) - Chrome DevTools Protocol
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [gorilla/mux](https://github.com/gorilla/mux) - HTTP router
