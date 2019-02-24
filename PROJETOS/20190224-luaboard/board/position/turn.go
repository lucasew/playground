package position

import (
    "math"
    "log"
)

// Positivo = Anti horário, Negativo = Horário => turned
func (p Position) PureTurn(angle, viewAngle Angle) Angle {
    rad := angle.Radian()
    view := viewAngle.Radian()
    if math.Abs(rad) > view {
        rad = (rad * (view/rad))
	}
	return NewAngleRadian(rad)
}

func (p *Position) DOTurn(angle, viewAngle Angle) Angle {
    log.Printf("CALL: (%v)DOTurn(%v)", p, angle.Degree())
	whatToTurn := p.PureTurn(angle, viewAngle)
	p.Heading.Value += whatToTurn.Value
    return p.GetHeading()
}

