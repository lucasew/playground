package position

import (
	"log"
	"math"
)

const oneDegree = 2.0 * math.Pi / 360

type Position struct {
	X, Y    float64
	Heading Angle
}

func (p Position) GetDistance(q Position) float64 {
	log.Printf("CALL: (%v)GetDistance(%v)\n", p, q)
	x := math.Abs(p.X - q.X)
	y := math.Abs(p.Y - q.Y)
	return math.Sqrt((x * x) + (y * y))
}

func (p *Position) GetHeading() Angle {
	log.Printf("CALL: (%v)GetHeading()", p)
	// Mantém tudo no intervalo de uma volta
	radians := p.Heading.Radian()
	for radians >= 2.0*math.Pi {
		radians -= 2.0 * math.Pi
	}
	for radians < 0 {
		radians += 2.0 * math.Pi
	}
	p.Heading = NewAngleRadian(radians)
	return p.Heading
}
