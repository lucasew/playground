package main

import (
	"fmt"
)

func main() {
	var A, B, C int
	fmt.Scanf("%d %d %d", &A, &B, &C)
	fmt.Printf("%d eh o maior\n", maior(maior(A, B), C))
}

func maior(a, b int) int {
	return (a + b + abs(a-b)) / 2
}

func abs(a int) int {
	if a <= 0 {
		return a * -1
	}
	return a
}
