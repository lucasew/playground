package board

import (	
    "github.com/lucasew/luaboard/board/position"
    "log"
)


type PlayerContext struct {
    Board *Board;
    Player *Player;
}

// Quantos graus eu preciso virar para ficar de frente com o ponto
func (pc *PlayerContext) GetDegreeTowardsPoint(x, y float64) position.Angle {
    log.Printf("CALL: GetDegreeTowardsPoint(%v, %v)\n", x, y)
    angleTowards := pc.Player.Position.GetAngleTowards(position.Position{X: x, Y: y})
    return position.NewAngleRadian(position.RadToDeg(pc.Player.Position.Heading.Radian() - angleTowards.Radian()))
}

func (pc *PlayerContext) GoAhead(distance float64) position.Position {
    log.Printf("CALL: GoAhead(%v)\n", distance)
    if distance > pc.Player.Speed {
        distance = pc.Player.Speed
    }
    destination := pc.Player.Position.DOGoAhead(distance)
    pc.Player.UseMana()
    return destination
}

