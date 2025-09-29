package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

func init() {
    apps["soma"] = NewSomaApp(42069)
}

type SomaApp struct {
    Port int
}

func NewSomaApp(port int) App {
    return SomaApp{
        Port: port,
    }
}

func (e SomaApp) Client(args []string) error {
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
    fmt.Fprintln(conn, "bye")
    time.Sleep(1*time.Second)
    return nil
}

func (e SomaApp) Server(args []string) error {
    ln, err := net.Listen("tcp", fmt.Sprintf(":%d", e.Port))
    if err != nil {
        return err
    }
    defer ln.Close()
    for {
        var soma float64 = 0
        conn, err := ln.Accept()
        if err != nil {
            return err
        }
        defer conn.Close()
        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            text := scanner.Text()
            if text == "bye" {
                break
            }
            num, err := strconv.ParseFloat(text, 64)
            if err != nil {
                fmt.Fprintln(conn, err.Error())
            }
            soma += num
        }
        fmt.Fprintf(conn, "%f\n", soma)
        fmt.Fprintln(conn)
    }
}
