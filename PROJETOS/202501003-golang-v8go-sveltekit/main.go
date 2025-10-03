package main

import (
	"github.com/davecgh/go-spew/spew"
	"rogchap.com/v8go"
)

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
}
