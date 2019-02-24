package position

import (
)
// Sempre será salvo como radiano
type Angle struct {
    Value float64
}
func (a Angle) Radian() float64 {
    return a.Value
}

func (a Angle) Degree() float64 {
    return RadToDeg(a.Value)
}

func NewAngleDegree(angle float64) Angle {
    return Angle {
        Value: DegToRad(angle),
    }
}

func NewAngleRadian(angle float64) Angle {
    return Angle {
        Value: angle,
    }
}

// DegToRad Converte os graus para radianos
func DegToRad(degree float64) float64 {
	return oneDegree * degree
}

func RadToDeg(rad float64) float64 {
	return rad / oneDegree
}
