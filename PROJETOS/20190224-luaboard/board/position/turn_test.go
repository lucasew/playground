package position

import (
    "testing"
    "math"
)

var viewAngle = NewAngleRadian(math.Pi/6) // 30 graus

func TestPositionPureTurnRadian(t *testing.T) {
    cases := []struct{
        Pos Position;
        Angle Angle;
        Expected Angle;
    }{
        {
            Pos: Position{Heading: NewAngleRadian(0)},
            Angle: NewAngleRadian(1.1*math.Pi/6),
            Expected: viewAngle,
        },
    }
    for _, cse := range(cases) {
        get := cse.Pos.PureTurn(cse.Angle, viewAngle)
        if (math.Abs(get.Radian() - cse.Expected.Radian()) > ErrorMargin) {
            t.Errorf("expected: %f, get: %f", cse.Expected.Radian(), get.Radian())
        }
    }
}
