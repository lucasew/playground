package position

import (
	"math"
	"testing"
)

type DegRadCase struct {
	Deg float64
	Rad float64
}

// Nota ao testador: A função Position.GetHeading não retorna 2.0*math.Pi, no cenário que ela retornaria, ela retorna 0 que é equivalente
var DegRadCases = []DegRadCase{
	{
		Rad: 0,
		Deg: 0.0,
	},
	{
		Rad: math.Pi / 2,
		Deg: 90.0,
	},
	{
		Rad: math.Pi,
		Deg: 180.0,
	},
	{
		Rad: 3.0 * math.Pi / 2.0,
		Deg: 270,
	},
}

func TestDegToRad(t *testing.T) {
	for _, cse := range DegRadCases {
		get := DegToRad(cse.Deg)
		if math.Abs(get-cse.Rad) > ErrorMargin {
			t.Errorf("Expected: %f, Get: %f", cse.Rad, get)
		}
	}
}

func TestRadToDeg(t *testing.T) {
	for _, cse := range DegRadCases {
		get := RadToDeg(cse.Rad)
		if math.Abs(get-cse.Deg) > ErrorMargin {
			t.Errorf("Expected: %f, Get: %f", cse.Deg, get)
		}
	}
}
