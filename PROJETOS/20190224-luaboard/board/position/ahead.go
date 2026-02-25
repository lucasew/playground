package position

import (
	"log"
	"math"
)

func (p Position) PureGoAhead(distance float64) Position {
	res := Position{}
	res.Heading = p.Heading
	res.X = p.X + distance*math.Cos(p.Heading.Radian())
	res.Y = p.Y + distance*math.Sin(p.Heading.Radian())
	return res
}

func (p *Position) DOGoAhead(distance float64) Position {
	log.Printf("CALL: DOGoAhead(%v)\n", distance)
	*p = p.PureGoAhead(distance)
	return *p
}

func (p Position) IsCanGoTowards(q Position, viewAngle Angle) bool {
	return math.Abs(p.GetAngleTowards(q).Radian()-p.GetHeading().Radian()) < viewAngle.Radian()/2
}
