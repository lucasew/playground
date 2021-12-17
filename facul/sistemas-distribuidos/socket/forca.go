package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
)

func init() {
    apps["forca"] = NewForcaApp(42069)
}

type ForcaApp struct {
    Port int
    Words []string
}

func NewForcaApp(port int) App {
    return ForcaApp{
        Port: port,
        Words: []string{
            "bolacha",
            "biscoito",
        },
    }
}

func (e *ForcaApp) GetWord() string {
    idx := rand.Intn(len(e.Words))
    return e.Words[idx]
}

func (e ForcaApp) Client(args []string) error {
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

func (e ForcaApp) Server(args []string) error {
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
        scanner := bufio.NewScanner(conn)
        word := e.GetWord()
        placeholder := []byte(strings.Repeat("_", len(word)))
        fmt.Fprintln(conn, "Pressione enter para iniciar")
        for {
            tentativas_restantes := 3
            _, err := fmt.Fprintln(conn, string(placeholder))
            if err != nil {
                log.Println(err.Error())
                break
            }
            _, err = fmt.Fprintf(conn, "Tentativas restantes: %d\n", tentativas_restantes)
            if err != nil {
                log.Println(err.Error())
                break
            }
            if (!scanner.Scan()) {
                break
            }
            text := scanner.Text()
            if text == "bye" {
                break
            }
            if len(text) == 1 {
                found := false
                end_of_game := true
                for i := 0; i < len(word); i++ {
                    if word[i] == text[0] {
                        found = true
                        placeholder[i] = text[0]
                    }
                    if placeholder[i] == '_' {
                        end_of_game = false
                    }
                }
                if found {
                    fmt.Fprintln(conn, "Você encontrou letras")
                }
                if end_of_game {
                    fmt.Fprintln(conn, "Você ganhou")
                    break
                }
                if !found {
                    if tentativas_restantes == 0 {
                        fmt.Fprintln(conn, "Você perdeu")
                        break
                    }
                    tentativas_restantes--
                }
            } else {
                fmt.Fprintln(conn, "Você passou algo maior que uma letra")
            }
        }
        conn.Close()
    }
}
