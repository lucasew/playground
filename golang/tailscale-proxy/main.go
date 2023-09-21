package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
    "net"
    "context"

	"tailscale.com/tsnet"
)

var remoteHost string
var remoteHostURL *url.URL
var enableFunnel bool
var tsHost string
var addr string

var cancel = func() {}
var ctx = context.Background()

func init() {
    var err error
    flag.StringVar(&remoteHost, "h", "", "Where to forward the connection")
    flag.BoolVar(&enableFunnel, "f", false, "Enable tailscale funnel")
    flag.StringVar(&tsHost, "n", "demoproxy", "Hostname in tailscale devices list")
    flag.StringVar(&addr, "addr", ":443", "Port to listen")
    flag.Parse()
    remoteHostURL, err = url.Parse(remoteHost)
    if err != nil {
        log.Fatal(err)
    }
    ctx, cancel = context.WithCancel(ctx)
}

func handleServer(ln net.Listener) {
    err := http.Serve(ln, nil)
    if err != nil {
        panic(err)
    }
    defer cancel()
    defer ln.Close()
}


func main() {
    handler := func(p *httputil.ReverseProxy) func (http.ResponseWriter, *http.Request) {
        return func(w http.ResponseWriter, r *http.Request) {
            // r.Host = remoteHostURL.Host
            p.ServeHTTP(w, r)
        }
    }
    proxy := httputil.NewSingleHostReverseProxy(remoteHostURL)
    http.HandleFunc("/", handler(proxy))
    s := new(tsnet.Server)
    s.Hostname = tsHost
    if enableFunnel {
        ln, err := s.ListenFunnel("tcp", addr)
        if err != nil {
            log.Fatal(err)
        }
        go handleServer(ln)
    } else {
        ln, err := s.Listen("tcp", addr)
        if err != nil {
            log.Fatal(err)
        }
        go handleServer(ln)
    }
    <-ctx.Done()
}
