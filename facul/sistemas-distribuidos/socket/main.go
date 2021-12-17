package main

import (
	"fmt"
	"os"
)

func main() {
    if len(os.Args) < 3 {
        panic("argumentos insuficientes: app [server,client] resto")
    }
    app := apps[os.Args[1]]
    command := os.Args[2]
    rest := []string{}
    if len(os.Args) > 3 {
        rest = os.Args[3:len(os.Args) - 1]
    }
    if app == nil {
        panic(fmt.Sprintf("app %s não encontrado", os.Args[1]))
    }
    var err error
    if command == "client" {
        err = app.Client(rest)
    } else if command == "server" {
        err = app.Server(rest)
    } else {
        panic(fmt.Sprintf("o comando %s não existe, apenas client e server", command))
    }
    if err != nil {
        panic(err)
    }
    fmt.Println("programa finalizado")
}
