package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func init() {
    apps["banco_conc"] = NewBancoConcApp(42069)
}

type BancoConcApp struct {
    sync.Mutex
    Port int
    saldos map[int]float64
}

func (a *BancoConcApp) Deposito(conta int, valor float64) {
    if valor <= 0 {
        return
    }
    a.Lock()
    defer a.Unlock()
    val, ok := a.saldos[conta]
    if !ok {
        val = 0
    }
    a.saldos[conta] = val + valor
}

func (a *BancoConcApp) Saldo(conta int) float64 {
    a.Lock()
    defer a.Unlock()
    val, ok := a.saldos[conta]
    if !ok {
        val = 0
    }
    return val
}
func (a *BancoConcApp) Saque(conta int, valor float64) {
    if valor <= 0 {
        return
    }
    a.Lock()
    defer a.Unlock()
    val, ok := a.saldos[conta]
    if !ok {
        val = 0
    }
    if val - valor < 0 {
        return
    }
    a.saldos[conta] = val - valor
}

func (a *BancoConcApp) Transferencia(conta int, outraConta int, valor float64) {
    if valor <= 0 {
        return
    }
    a.Lock()
    defer a.Unlock()
    va, ok := a.saldos[conta]
    if !ok {
        va = 0
    }
    vb, ok := a.saldos[outraConta]
    if !ok {
        vb = 0
    }
    if va - valor < 0 {
        return
    }
    a.saldos[conta] = va - valor
    a.saldos[outraConta] = vb + valor
}


func NewBancoConcApp(port int) App {
    return &BancoConcApp{
        Port: port,
        saldos: map[int]float64{},
    }
}

func (a *BancoConcApp) Client(args []string) error {
    scanner := bufio.NewScanner(os.Stdin)
    conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", a.Port))
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

func (a *BancoConcApp) Server(args []string) error {
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
            defer conn.Close()
            conta := 0
            scanner := bufio.NewScanner(conn)
            for scanner.Scan() {
                text := scanner.Text()
                parts := strings.Split(text, " ")
                if len(parts) == 0 {
                    fmt.Fprintln(conn, "warn: linha vazia enviada")
                }
                switch (parts[0]) {
                    case "b":
                        return
                    case "l":
                        if len(parts) == 2 {
                            acc, err := strconv.ParseInt(parts[1], 10, 32)
                            if err != nil {
                                fmt.Fprintln(conn, "erro: número de conta inválido")
                            }
                            conta = int(acc)
                        } else {
                            fmt.Fprintln(conn, "l[ogin] num_conta")
                        }
                    case "d":
                        if len(parts) == 2 {
                            acc, err := strconv.ParseFloat(parts[1], 32)
                            if err != nil {
                                fmt.Fprintln(conn, "erro: valor inválido")
                            }
                            a.Deposito(conta, acc)
                        } else {
                            fmt.Fprintln(conn, "d[deposit] valor")
                        }
                    case "s":
                        if len(parts) == 2 {
                            acc, err := strconv.ParseFloat(parts[1], 32)
                            if err != nil {
                                fmt.Fprintln(conn, "erro: valor inválido")
                            }
                            a.Saque(conta, acc)
                        } else {
                            fmt.Fprintln(conn, "s[aque] valor")
                        }
                    case "m":
                        fmt.Fprintln(conn, fmt.Sprintf("Saldo: %.2f", a.Saldo(conta)))
                    case "t":
                        if len(parts) == 3 {
                            acc, err := strconv.ParseFloat(parts[1], 32)
                            if err != nil {
                                fmt.Fprintln(conn, "erro: valor inválido")
                            }
                            a.Saque(conta, acc)
                            outraConta, err := strconv.ParseInt(parts[2], 10, 32)
                            if err != nil {
                                fmt.Fprintln(conn, "erro: número de conta inválido")
                            }
                            a.Transferencia(conta, int(outraConta), acc)
                        } else {
                            fmt.Fprintln(conn, "t[ransfer] valor conta")
                        }
                    default:
                        fmt.Fprintln(conn, "Comandos disponíveis: b l d s t m")
                }
            }
        })(conn)
    }
}
