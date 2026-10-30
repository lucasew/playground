// TORA opens the operations-research page in the system web view.
//
//	go run github.com/lewtec/lewkit/cmd/lewkit release run --config eletrocromo.json
package main

import (
	"context"

	"github.com/lewtec/lewkit/x/app"
	"github.com/lewtec/lewkit/x/entry"

	"github.com/lucasew/tora"
)

func init() { entry.Bind(runApp) }

func main() { entry.Main(context.Background(), runApp) }

func runApp(ctx context.Context) error {
	handler, err := tora.Handler(ctx)
	if err != nil {
		return err
	}
	return app.App{
		Title:   "TORA",
		Width:   1060,
		Height:  780,
		Handler: app.Web(handler),
	}.Run(ctx)
}
