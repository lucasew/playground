package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sqlc-atlas/db"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	"github.com/go-json-experiment/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var DB_URL string

func init() {
	flag.StringVar(&DB_URL, "db", os.Getenv("DATABASE_URL"), "Postgres")
	flag.Parse()
}

func getDatabase(ctx context.Context) db.DBTX {
	conn, err := pgx.Connect(ctx, DB_URL)
	if err != nil {
		panic(err)
	}
	return conn
}

func runMigrations() {
	database_url, err := url.Parse(DB_URL)
	if err != nil {
		panic(err)
	}

	mate := dbmate.New(database_url)
	mate.FS = migrationsFS
	if err := mate.CreateAndMigrate(); err != nil {
		panic(err)
	}
}

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

func main() {
	runMigrations()
	database := db.New(getDatabase(context.Background()))
	http.HandleFunc("/create-user", func(w http.ResponseWriter, r *http.Request) {
		userName := r.FormValue("name")
		if userName == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "bad user name passed")
			return
		}
		entity, err := database.CreateUser(r.Context(), pgtype.Text{String: userName, Valid: true})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "invalid insertion: %s", err.Error())
			return
		}
		json.MarshalWrite(w, entity)

	})
	http.HandleFunc("/audit", func(w http.ResponseWriter, r *http.Request) {
		items, err := database.GetUserAudits(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "problem in database: %s", err.Error())
			return
		}
		json.MarshalWrite(w, items)
	})
	log.Printf("Escutando :5090")
	http.ListenAndServe(":5090", nil)
	log.Printf("Parando...")
}
