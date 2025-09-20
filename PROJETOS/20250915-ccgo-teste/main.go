// Code generated for linux/amd64 by 'ccgo main.c', DO NOT EDIT.

//go:build linux && amd64

package main

import (
	"reflect"
	"unsafe"

	"modernc.org/libc"
)

var _ reflect.Type
var _ unsafe.Pointer

func main1(tls *libc.TLS, argc int32, argv uintptr) (r int32) {
	libc.Xprintf(tls, __ccgo_ts, 0)
	return r
}

func main() {
	libc.Start(main1)
}

var __ccgo_ts = (*reflect.StringHeader)(unsafe.Pointer(&__ccgo_ts1)).Data

var __ccgo_ts1 = "Hello, world\n\x00"
