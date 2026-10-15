# Build stage
FROM golang:1.27-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o rendereiro cmd/rendereiro/main.go

# Runtime stage
FROM chromedp/headless-shell:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy binary and web assets from builder
COPY --from=builder /app/rendereiro .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config.yaml .

# Create profiles directory
RUN mkdir -p /app/profiles

# Expose port
EXPOSE 8080

# Run the application
CMD ["./rendereiro", "serve"]
