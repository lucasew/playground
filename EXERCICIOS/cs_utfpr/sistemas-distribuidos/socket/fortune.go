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
	apps["fortune"] = NewFortuneApp(42069)
}

type FortuneApp struct {
	Port     int
	Fortunes []string
}

func NewFortuneApp(port int) App {
	return &FortuneApp{
		Port: port,
		Fortunes: []string{
			"2+2 = 4",
			"é melhor morrer do que perder a vida",
			"ssdfjsjflsakjf",
		},
	}
}

func (e *FortuneApp) GetFortune() string {
	idx := rand.Intn(len(e.Fortunes))
	return e.Fortunes[idx]
}

func (e *FortuneApp) SetFortune(msg string) {
	e.Fortunes = append(e.Fortunes, msg)
}

func (e *FortuneApp) Client(args []string) error {
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

func (e *FortuneApp) Server(args []string) error {
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
				if strings.ToUpper(text) == "GET-FORTUNE" {
					_, err = fmt.Fprintln(conn, e.GetFortune())
				}
				if strings.ToUpper(text) == "SET-FORTUNE" {
					if !scanner.Scan() {
						return
					}
					e.SetFortune(scanner.Text())
				}
				if err != nil {
					log.Printf("err: %s\n", err.Error())
					return
				}
			}
			defer conn.Close()
		})(conn)
	}
}
