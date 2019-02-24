package position

import (
    "testing"
    "math"
)

func TestPositionGetAngleTowards(t *testing.T) {
    cases := []struct{
        P, Q Position;
        Res Angle;
    }{
        {
            P: Position{X: 0, Y: 0, Heading: NewAngleRadian(0)},
            Q: Position{X: 3, Y: 4}, // Triângulo retângulo
            Res: NewAngleRadian(0.927295),
        },
        {
            P: Position{X: 0, Y: 0, Heading: NewAngleRadian(0)},
            Q: Position{X: 233, Y: -223}, // Valor aleatório pra fazer na mão e ver se bate
            Res: NewAngleRadian(-0.763471853),
        },
    }
    for _, cse := range(cases) {
        get := cse.P.GetAngleTowards(cse.Q)
        if (math.Abs(get.Radian() - cse.Res.Radian()) > ErrorMargin) {
            t.Errorf("Expected: %f, get: %f", cse.Res, get)
        }
    }
}
