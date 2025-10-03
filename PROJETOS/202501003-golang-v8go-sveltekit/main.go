package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"rogchap.com/v8go"
)

//go:embed all:sveltekit-frontend/.svelte-kit/output
var outputData embed.FS

type SvelteKitServer struct {
	assets  fs.FS
	isolate *v8go.Isolate
	bundle  *v8go.UnboundScript
}

// ServeHTTP implements http.Handler.
func (s *SvelteKitServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assetFS := MustSubFS(s.assets, "client")
	staticFile, err := assetFS.Open(r.URL.Path)
	if staticFile != nil {
		defer staticFile.Close()
	}
	if err == nil {
		io.Copy(w, staticFile)
	}
	ctx := v8go.NewContext(s.isolate)
	defer ctx.Close()
	err = s.fillIsolate(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		spew.Dump(err)
		return
	}
	spew.Dump(s.bundle.Run(ctx))
	w.WriteHeader(404)
}

func (s *SvelteKitServer) fillIsolate(ctx *v8go.Context) error {
	global := ctx.Global()
	consoleObj, err := v8go.NewObjectTemplate(s.isolate).NewInstance(ctx)
	if err != nil {
		return fmt.Errorf("v8-polyfills/console: %w", err)
	}

	logFn := v8go.NewFunctionTemplate(s.isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		spew.Dump(info.Args())
		return nil
	})
	if err := consoleObj.Set("log", logFn.GetFunction(ctx)); err != nil {
		return fmt.Errorf("v8-polyfills/console: %w", err)
	}
	if err := global.Set("console", consoleObj); err != nil {
		return fmt.Errorf("v8-polyfills/console: %w", err)
	}
	return nil
}

func NewSvelteKitServer(assets fs.FS) (*SvelteKitServer, error) {
	iso := v8go.NewIsolate()

	bundleFile, err := assets.Open("server/bundle.js")
	if err != nil {
		return nil, err
	}
	bundleContent, err := io.ReadAll(bundleFile)
	if err != nil {
		return nil, err
	}

	bundleCompiled, err := iso.CompileUnboundScript(string(bundleContent), "index.js", v8go.CompileOptions{})
	if err != nil {
		return nil, err
	}

	return &SvelteKitServer{
		assets:  assets,
		isolate: iso,
		bundle:  bundleCompiled,
	}, nil
}

func MustSubFS(superset fs.FS, subpath string) fs.FS {
	ret, err := fs.Sub(superset, subpath)
	if err != nil {
		panic(err)
	}
	return ret
}

func main() {
	iso := v8go.NewIsolate()
	defer iso.Dispose()
	ctx := v8go.NewContext(iso)
	defer ctx.Close()
	val, err := ctx.RunScript("2", "main.js")
	if err != nil {
		panic(err)
	}
	spew.Dump(val)
	server, err := NewSvelteKitServer(MustSubFS(outputData, "sveltekit-frontend/.svelte-kit/output"))
	if err != nil {
		panic(err)
	}
	http.ListenAndServe(":5555", server)
}
