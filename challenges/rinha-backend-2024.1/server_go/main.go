package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tigerbeetle_go "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	// "github.com/tigerbeetle/tigerbeetle-go"
	// "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

var (
	httpAddr        string
	tigerbeetleHost string
)

const (
	TIGERBETTLE_MAX_CONCURRENCY = 256
)

func init() {
	portFromEnv := os.Getenv("PORT")
	if portFromEnv == "" {
		portFromEnv = "3001"
	}
	flag.StringVar(&httpAddr, "addr", fmt.Sprintf(":%s", portFromEnv), "Address where to listen for the server")

	flag.StringVar(&tigerbeetleHost, "t", os.Getenv("TB_ADDRESS"), "How to connect to tigerbeetle")
	flag.Parse()
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("can't prepare server: %s", err)
	}
	defer app.Close()
	log.Printf("Starting http server on %s", httpAddr)
	err = http.ListenAndServe(httpAddr, app)
	if err != nil {
		log.Fatalf("can't start http server: %s", err)
	}
}

type App struct {
	tigerbeetle tigerbeetle_go.Client
}

func NewApp() (*App, error) {
	client, err := tigerbeetle_go.NewClient(types.ToUint128(0), strings.Split(tigerbeetleHost, ","), TIGERBETTLE_MAX_CONCURRENCY)
	if err != nil {
		return nil, err
	}
	return &App{tigerbeetle: client}, nil
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "foi!")
}

func (a *App) Close() error {
	a.tigerbeetle.Close()
	return nil
}
