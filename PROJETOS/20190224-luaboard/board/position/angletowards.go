package position

import (
	"log"
	"math"
)

func (p Position) GetAngleTowards(q Position) Angle {
	log.Printf("CALL: (%v)GetAngleTowards(%v)\n", p, q)
	x := p.X - q.X
	y := p.Y - q.Y
	return NewAngleRadian(math.Atan(y / x))
}
