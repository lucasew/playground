package position

import (
	"math"
	"testing"
)

const ErrorMargin = 0.0001

func TestPositionGetDistance(t *testing.T) {
	cases := []struct {
		ReqA, ReqB Position
		Res        float64
	}{
		{
			ReqA: Position{X: 0, Y: 0},
			ReqB: Position{X: 2, Y: 2},
			Res:  2 * math.Sqrt(2),
		},
	}
	for _, cse := range cases {
		get := cse.ReqA.GetDistance(cse.ReqB)
		if math.Abs(get-cse.Res) > ErrorMargin {
			t.Errorf("Expected: %f Get: %f", cse.Res, get)
		}
	}
}

func TestPositionGetHeading(t *testing.T) {
	cases := []struct {
		Req Position;
		Res Angle;
	}{
		{
			Req: Position{Heading: NewAngleRadian(3.0 * math.Pi)},
			Res: NewAngleRadian(math.Pi),
		},
		{
			Req: Position{Heading: NewAngleRadian(4.0 * math.Pi)},
			Res: NewAngleRadian(0),
		},
		{
			Req: Position{Heading: NewAngleRadian(math.Pi)},
			Res: NewAngleRadian(math.Pi),
		},
	}
	for _, cse := range cases {
		get := cse.Req.GetHeading().Radian()
		if math.Abs(get - cse.Res.Radian()) > ErrorMargin {
			t.Errorf("Expected: %f Get: %f", cse.Res.Radian(), get)
		}
	}
}

