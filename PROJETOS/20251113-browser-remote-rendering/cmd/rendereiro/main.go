package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lewtec/rendereiro/internal/app"
	"github.com/lewtec/rendereiro/internal/config"
	"github.com/spf13/cobra"
)

var (
	configPath string
	stateDir   string
	port       int
	host       string
	browserPathFlag string
	verboseFlag bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "rendereiro",
		Short: "Rendereiro - Remote Browser Service via SSE",
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the remote browser server",
		Run:   runServe,
	}

	serveCmd.Flags().StringVar(&configPath, "config", "config.yaml", "Path to config.yaml")
	serveCmd.Flags().StringVar(&stateDir, "state", "", "Directory for Chrome profiles (default: <config_dir>/profiles)")
	serveCmd.Flags().IntVar(&port, "port", 8080, "HTTP server port")
	serveCmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host to bind to")
	serveCmd.Flags().StringVar(&browserPathFlag, "browser-path", "", "Path to the Chromium executable (overrides config and env var)")
	serveCmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Enable verbose logging (DEBUG level)")

	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runServe(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set default state directory if not provided
	if stateDir == "" {
		stateDir = "./profiles"
	}

	// Create app
	appInputs := app.AppInputs{
		StateDir: stateDir,
		Config:   cfg,
		Port:     port,
		Host:     host,
		Verbose:  verboseFlag,
	}

	// Determine final browser path
	if browserPathFlag != "" {
		appInputs.BrowserPath = browserPathFlag
	} else if cfg.BrowserPath != "" {
		appInputs.BrowserPath = cfg.BrowserPath
	}

	application, err := app.New(appInputs)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	defer application.Cleanup()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", host, port)
	server := &http.Server{
		Addr:         addr,
		Handler:      application,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
