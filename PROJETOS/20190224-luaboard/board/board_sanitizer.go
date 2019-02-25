package board

import (
	"log"
)

func (b *Board) BoardSanitizer() {
	log.Println("SANITIZER: Colocando as coisas no lugar")
	for _, player := range b.Players {
		if player.Position.X > float64(b.SizeX) {
			player.Position.X = float64(b.SizeX)
		} else if player.Position.X < 0 {
			player.Position.X = 0
		}
		if player.Position.Y > float64(b.SizeY) {
			player.Position.Y = float64(b.SizeY)
		} else if player.Position.Y < 0 {
			player.Position.Y = 0
		}
	}
}
