package main

import (
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "works")
}

func main() {
	// cool, this actually works!
	err := http.ListenAndServe("127.0.0.69:42070", http.HandlerFunc(Handler))
	if err != nil {
		panic(err)
	}
}
