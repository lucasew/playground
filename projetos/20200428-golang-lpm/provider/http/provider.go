package provider

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	definition "github.com/lucasew/lpm/lib"
)

type Provider struct {
	definition.Provider
}

func (ctx Provider) Provide() error {
	resp, err := http.Get(ctx.Namespace.ScopedPath())
	if err != nil {
		return err
	}
	outfile := ctx.DataFolder
	for _, elem := range ctx.Namespace.ScopedSlices() {
		outfile = filepath.Join(outfile, elem)
	}
	println(outfile)
	stat, err := os.Stat(outfile)
	if os.IsNotExist(err) {
		goto download
	}
	if stat.Size() != resp.ContentLength && resp.ContentLength != -1 {
		goto download
	}
	return nil
download:
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
