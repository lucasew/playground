package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func init() {
    apps["eco"] = NewEcoApp(42069)
}

type EcoApp struct {
    Port int
}

func NewEcoApp(port int) App {
    return EcoApp{
        Port: port,
    }
}

func (e EcoApp) Client(args []string) error {
    scanner := bufio.NewScanner(os.Stdin)
    conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", e.Port))
    if err != nil {
        return err
    }
    defer conn.Close()
    go func() {
        connScanner := bufio.NewScanner(conn)
        for connScanner.Scan() {
            text := connScanner.Text()
            fmt.Println(text)
        }
        log.Printf("Recebido dado do servidor")
    }()
    for scanner.Scan() {
        text := scanner.Text()
        _, err := fmt.Fprintln(conn, text)
        if err != nil {
            return err
        }
    }
    return nil
}

func (e EcoApp) Server(args []string) error {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", e.Port))
    if err != nil {
        return err
    }
    defer ln.Close()
    for {
        conn, err := ln.Accept()
        if err != nil {
            return err
        }
        go (func(conn net.Conn) {
            scanner := bufio.NewScanner(conn)
            for scanner.Scan() {
                text := scanner.Text()
                _, err := fmt.Fprintln(conn, text)
                if err != nil {
                    log.Printf("err: %s\n", err.Error())
                    return
                }
            }
            defer conn.Close()
        })(conn)
    }
}
