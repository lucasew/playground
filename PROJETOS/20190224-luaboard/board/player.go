package board

import (
	"github.com/lucasew/luaboard/board/position"
	"log"
	"time"
)

type Player struct {
	AttackPoints  int
	DefensePoints int
	Speed         float64
	Mana          int
	Position      position.Position
}

// TODO: Algum critério pro tempo
func (p *Player) GetManaIncreaseInterval() time.Duration {
	return time.Millisecond * 500
}

func (p *Player) AddMana() {
	p.Mana++
}

func (p *Player) UseMana() {
	log.Printf("CALL: UseMana\n")
	if p.Mana <= 0 {
		time.Sleep(p.GetManaIncreaseInterval())
	}
	p.Mana--
}
