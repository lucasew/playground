// Package milhao3 serves Show do Milhão 3 from the embedded zip.
// The archive is opened in place. Lewkit's SPA middleware serves a real
// file when the path names one, and index.html for any other GET.
package milhao3

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"

	lewfs "github.com/lewtec/lewkit/x/fs"
	"github.com/lewtec/lewkit/x/fs/zip"
	"github.com/lewtec/lewkit/x/http/middleware"
)

//go:embed milhao3.zip
var gameZip []byte

// Handler serves the game. The zip's single top directory is the site root.
func Handler(ctx context.Context) (http.Handler, error) {
	archive, err := zip.Open(ctx, bytes.NewReader(gameZip))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	root, err := lewfs.StripTopDirectory(archive)
	if err != nil {
		return nil, fmt.Errorf("site root: %w", err)
	}
	// Deflated zip members are streams. middleware.SPA serves through
	// http.ServeFileFS, which answers Range only when the file can seek.
	// Video and audio elements send Range. Buffer those members.
	return middleware.SPA(seekFS{base: root}, nil), nil
}

// seekFS opens files that implement io.Seeker.
// A directory is returned as stored. A file that already seeks is returned
// as stored. Any other file is read once into memory.
type seekFS struct {
	base fs.FS
}

func (s seekFS) Open(name string) (fs.File, error) {
	file, err := s.base.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if info.IsDir() {
		return file, nil
	}
	if _, ok := file.(io.Seeker); ok {
		return file, nil
	}
	data, err := io.ReadAll(file)
	closeErr := file.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return &memFile{info: info, Reader: bytes.NewReader(data)}, nil
}

func (s seekFS) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(s.base, name)
}

func (s seekFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(s.base, name)
}

// memFile is one buffered zip member.
type memFile struct {
	info fs.FileInfo
	*bytes.Reader
}

func (f *memFile) Stat() (fs.FileInfo, error) { return f.info, nil }

func (f *memFile) Close() error { return nil }
