package main

import (
	"fmt"
)

func main() {
	var rendimento, gasto float32
	var distancia int
	fmt.Scanf("%d\n%f", &distancia, &gasto)
	rendimento = float32(distancia) / gasto
	fmt.Printf("%.3f km/l\n", rendimento)
}
