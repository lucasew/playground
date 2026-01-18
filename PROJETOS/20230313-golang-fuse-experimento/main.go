package main

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"
)

func main() {
	var mountpoint string
	flag.StringVar(&mountpoint, "m", "/tmp/mount-test", "Where to mount the test mountpoint")
	flag.Parse()
	c, err := fuse.Mount(mountpoint, fuse.FSName("demofs"), fuse.Subtype("awesome"))
	if err != nil {
		panic(err)
	}
	defer c.Close()
	err = fs.Serve(c, FS{})
	if err != nil {
		panic(err)
	}
}

var data = map[string]string{
	"a": "AA",
	"b": "BB",
}

var entities = []fuse.Dirent{}

func init() {
	var inode uint64 = 2
	for k := range data {
		entities = append(entities, fuse.Dirent{
			Inode: inode,
			Name:  k,
			Type:  fuse.DT_File,
		})
		inode++
	}
}

type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

type Dir struct{}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0o777
	return nil
}

func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for _, ent := range entities {
		if name == ent.Name {
			return file{ent: ent}, nil
		}
	}
	return nil, syscall.ENOENT
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return entities, nil
}

type file struct {
	ent fuse.Dirent
}

func (f file) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.ent.Inode
	a.Mode = 0o777
	a.Size = uint64(len(data[f.ent.Name]))
	return nil
}

func (f file) ReadAll(ctx context.Context) ([]byte, error) {
	fmt.Println("readall")
	return []byte(data[f.ent.Name]), nil
}
