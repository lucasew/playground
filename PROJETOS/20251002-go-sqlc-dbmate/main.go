package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sqlc-atlas/db"

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

func main() {
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
	log.Printf("Escutando :5090")
	http.ListenAndServe(":5090", nil)
}
