package main

import (
	"os"
	"time"
)

func main() {
	letra := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(letra)
		if err != nil && n != 1 {
			break
		}
		os.Stdout.Write(letra)
		time.Sleep(50 * time.Millisecond)
	}
}
