package main

import "fmt"

type Pessoa struct {
	Nome  string
	Idade int
}

func (p Pessoa) MostrarInfo() {
	fmt.Printf("%s tem %d anos!\n", p.Nome, p.Idade)
}

func main() {
	pessoas := []Pessoa{{
		Nome:  "Lucas",
		Idade: 18,
	},
		{
			Nome:  "João",
			Idade: 17,
		}}
	for _, pessoa := range pessoas { // Indice e valor
		pessoa.MostrarInfo()
	}
}
