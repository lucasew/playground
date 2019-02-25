package board

import (
	"github.com/lucasew/luaboard/board/position"
	"log"
	"math"
	"time"
)

type Board struct {
	SizeX, SizeY int
	Players      []*Player
	IsEnd        bool
}

func NewBoard(sizex, sizey int) *Board {
	log.Printf("Criando tabuleiro %v x %v...\n", sizex, sizey)
	return &Board{
		SizeX:   sizex,
		SizeY:   sizey,
		Players: []*Player{},
		IsEnd:   false,
	}
}

func (b *Board) StartSanitizerScheduler(d time.Duration) {
	sanitizer := func(b *Board) {
		for !b.IsEnd {
			b.BoardSanitizer()
			time.Sleep(d)
		}
	}
	go sanitizer(b)
}

func (b *Board) StartManaTicker(d time.Duration) {
	log.Printf("Iniciando contador de mana...\n")
	thicker := func() {
		for !b.IsEnd {
			for k, _ := range b.Players {
				b.Players[k].AddMana()
			}
			time.Sleep(d)
		}
	}
	go thicker()
}

func (b *Board) NewPlayer(atk, def int, speed float64) *PlayerContext {
	log.Printf("Criando jogador...\n")
	p := &Player{
		AttackPoints:  atk,
		DefensePoints: def,
		Speed:         speed,
		Mana:          0,
	}
	b.Players = append(b.Players, p)
	board := &PlayerContext{
		Board:  b,
		Player: p,
	}
	return board
}

func (b *Board) InitBoard() {
	log.Printf("Inicializando tabuleiro...\n")
	center := position.Position{
		X: float64(b.SizeX) / 2.0,
		Y: float64(b.SizeY) / 2.0,
	}
	var boardSlice float64 = 2 * math.Pi / float64(len(b.Players))
	for i := 0; i < len(b.Players); i++ {
		center.Heading = position.NewAngleRadian(boardSlice * float64(i))
		b.Players[i].Position = center.PureGoAhead(2)  // Distancia do centro do tabuleiro
		b.Players[i].Position.Heading = center.Heading // Virado para fora
	}
}
