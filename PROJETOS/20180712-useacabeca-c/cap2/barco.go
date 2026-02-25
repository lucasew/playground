package main

import "fmt"

func go_south_east(lat, lon *int) {
	*lat = *lat - 1
	*lon = *lon + 1
}

func main() {
	var lat int = 32
	var lon int = -64
	go_south_east(&lat, &lon)
	fmt.Printf("Estamos agora em (%d, %d)!\n", lat, lon)
}
