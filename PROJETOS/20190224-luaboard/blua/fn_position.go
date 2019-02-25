package blua

import (
	"github.com/lucasew/luaboard/board/position"
	"github.com/yuin/gopher-lua"
)

func LoadPosition(lw *LuaWrapper) {
	tbl := lw.State.NewTable()
	lw.State.SetField(tbl, "get_distance_to", func(l *lua.LState) int {
		x := float64(l.ToNumber(1))
		y := float64(l.ToNumber(2))
		destination := position.Position{
			X: x,
			Y: y,
		}
		distance := lw.Context.Player.Position.GetDistance(destination)
		lw.Context.Player.UseMana()
		l.Push(lua.LNumber(distance))
		return 1
	})
	lw.State.SetGlobal("position", tbl)
}
