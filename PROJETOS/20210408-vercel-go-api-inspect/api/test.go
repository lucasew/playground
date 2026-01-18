package handler

import (
	"net/http"

	"github.com/davecgh/go-spew/spew"
)

type Ret struct {
	Sr *http.Request
	Sw http.ResponseWriter
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	ret := Ret{
		Sr: r,
		Sw: w,
	}
	spew.Fdump(w, ret)
	return
}
