package main

import (
	"fmt"
	"strings"
)

var tracks = []string{
	"I left my  hearth in Harvard Med School",
	"Newark, Newark - a wonderful town",
	"Dancing with a Dork",
	"From here to maternity",
	"The girl from Iwo Jima",
}

func find_track(query string) {
	for i := 0; i < 5; i++ {
		if strings.Index(tracks[i], query) != -1 {
			fmt.Printf("Found: %s\n", tracks[i])
		}
	}
}

func main() {
	fmt.Printf("Pesquisar por: ")
	var query string
	fmt.Scan(&query)
	find_track(query)
}
