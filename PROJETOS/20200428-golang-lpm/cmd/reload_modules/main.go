package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	definition "github.com/lucasew/lpm/lib"
)

type Provider struct {
	definition.Provider
}

func (ctx Provider) Provide() error {
	basepath, err := os.Getwd()
	if err != nil {
		return err
	}
	flag.StringVar(&basepath, "b", basepath, "Plugin root path")
	flag.CommandLine.Parse(ctx.Args)
	stat, err := os.Stat(basepath)
	if err != nil {
		return ctx.Errorf(err.Error())
	}
	if !stat.IsDir() {
		return ctx.Errorf("plugin root (%s) need to be a directory", basepath)
	}
	pluginPaths, err := findGolangFolders(basepath)
	fmt.Printf("%+v", pluginPaths)
	return err
}

func findGolangFolders(base string) (folders []string, err error) {
	absbase, err := filepath.Abs(base)
	if err != nil {
		return
	}
	i := 0
	foldersMap := map[string]interface{}{} // remove duplicates the stonks way
	err = filepath.Walk(absbase, func(path string, info os.FileInfo, err error) error {
		defer func() { i++ }()
		if i == 0 {
			return nil
		}
		abspath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			foldersMap[filepath.Dir(abspath)] = nil
		}
		if info.IsDir() {
			childFolders, err := findGolangFolders(abspath)
			if err != nil {
				return err
			}
			for _, f := range childFolders {
				foldersMap[f] = nil
			}
		}
		return nil
	})
	folders = make([]string, 0, len(foldersMap))
	for k := range foldersMap {
		folders = append(folders, k)
	}
	foldersMap = nil // destroy the map
	return
}

func main() {
	pro := Provider{definition.Provider{
		Namespace:  definition.NewNamespace("internal/loader"),
		Args:       []string{"-b", "../../provider"},
		DataFolder: "../../data",
	}}
	err := pro.Provide()
	if err != nil {
		panic(err)
	}
}
