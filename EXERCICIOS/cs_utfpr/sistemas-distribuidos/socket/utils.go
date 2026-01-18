package main

type App interface {
	Client(args []string) error
	Server(args []string) error
}

var apps = map[string]App{}
