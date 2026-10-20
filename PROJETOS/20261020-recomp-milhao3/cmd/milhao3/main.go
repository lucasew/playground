// Milhao3 opens Show do Milhão 3 in the system web view.
//
//	go run github.com/lewtec/lewkit/cmd/lewkit release run --config eletrocromo.json
package main

import (
	"context"

	"github.com/lewtec/lewkit/x/app"
	"github.com/lewtec/lewkit/x/entry"

	"github.com/lucasew/milhao3"
)

func init() { entry.Bind(runApp) }

func main() { entry.Main(context.Background(), runApp) }

func runApp(ctx context.Context) error {
	handler, err := milhao3.Handler(ctx)
	if err != nil {
		return err
	}
	return app.App{
		Title:   "Show do Milhão",
		Width:   960,
		Height:  720,
		Handler: app.Web(handler),
	}.Run(ctx)
}
