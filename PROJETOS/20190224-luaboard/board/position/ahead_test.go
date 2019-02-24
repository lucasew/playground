package position

import (
    "math"
    "testing"
)

func TestPositionGoAhead(t *testing.T) {
	cases := []struct {
		Req      Position
		Distance float64
		Res      Position
	}{
		{
			Req:      Position{X: 10, Y: 10, Heading: NewAngleRadian(math.Pi)},
			Distance: 10,
			Res:      Position{X: 0, Y: 10, Heading: NewAngleRadian(math.Pi)},
		},
	}
	for _, cse := range cases {
		get := cse.Req.PureGoAhead(cse.Distance)
		if math.Abs(get.X-cse.Res.X) > ErrorMargin || math.Abs(get.Y-cse.Res.Y) > ErrorMargin {
			t.Errorf("Expected (%f, %f), Get: (%f, %f)", cse.Res.X, cse.Res.Y, get.X, get.Y)
		}
	}
}

func TestPositionIsCanGoTowards(t *testing.T) {
	viewAngle := NewAngleDegree(30.0)
	cases := []struct {
		ReqA, ReqB Position
		Res        bool
	}{
		{
			// atan de 1 (45)
			ReqA: Position{X: 0, Y: 0, Heading: NewAngleDegree(0)},
			ReqB: Position{X: 2, Y: 2},
			Res:  false,
		},
		{
			// atan de 0.25 (14 e pouco)
			ReqA: Position{X: 0, Y: 0, Heading: NewAngleDegree(0)},
			ReqB: Position{X: 8, Y: 2},
			Res:  true,
		},
		{
			// atan de 1 (45)
			ReqA: Position{X: 0, Y: 0, Heading: NewAngleDegree(31)},
			ReqB: Position{X: 2, Y: 2},
			Res:  true,
		},
	}
	for _, cse := range cases {
		get := cse.ReqA.IsCanGoTowards(cse.ReqB, viewAngle)
		if get != cse.Res {
			t.Errorf("Expected: %v, Get: %v", cse.Res, get)
		}
	}
}
