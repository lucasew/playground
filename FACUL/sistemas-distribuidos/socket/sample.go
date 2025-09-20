package main

import (
	"fmt"
	"net"
)

func init() {
    apps["sample"] = NewSampleApp(42069)
}

type SampleApp struct {
    Port int
}

func NewSampleApp(port int) App {
    return SampleApp{
        Port: port,
    }
}

func (a SampleApp) Client(args []string) error {
    conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", a.Port))
    if err != nil {
        return err
    }
    defer conn.Close()
    return nil
}

func (a SampleApp) Server(args []string) error {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", a.Port))
    if err != nil {
        return err
    }
    for {
        conn, err := ln.Accept()
        if err != nil {
            return err
        }
        go (func(conn net.Conn) {
            fmt.Fprintf(conn, "hello, %s", conn.RemoteAddr().String())
            defer conn.Close()
        })(conn)
    }
}
