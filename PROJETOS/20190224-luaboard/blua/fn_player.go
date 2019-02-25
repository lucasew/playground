package blua

import (
	"github.com/lucasew/luaboard/board/position"
	"github.com/yuin/gopher-lua"
)

func LoadPlayer(lw *LuaWrapper) {
	tbl := lw.State.NewTable()
	lw.State.SetField(tbl, "get_x", func(l *lua.LState) int {
		res := lw.Context.Player.Position.X
		l.Push(lua.LNumber(res))
		return 1
	})
	lw.State.SetField(tbl, "get_y", func(l *lua.LState) int {
		res := lw.Context.Player.Position.Y
		l.Push(lua.LNumber(res))
		return 1
	})
	lw.State.SetField(tbl, "get_heading", func(l *lua.LState) int {
		res := lw.Context.Player.Position.Heading.Degree()
		l.Push(lua.LNumber(res))
		return 1
	})
	lw.State.SetField(tbl, "get_pos", func(l *lua.LState) int {
		resx := lw.Context.Player.Position.X
		l.Push(lua.LNumber(resx))
		resy := lw.Context.Player.Position.Y
		l.Push(lua.LNumber(resy))
		return 2
	})
	// TODO: Arrumar essa parte
	lw.State.SetField(tbl, "turn_relative", func(l *lua.LState) int {
		angle := float64(l.ToNumber(1))
		turned := lw.Context.Player.TurnRelative(position.NewAngleDegree(angle), lw.ViewAngle)
		lw.Context.Player.UseMana()
		l.Push(lua.LNumber(turned))
		return 1
	})
	// TODO: Arrumar essa parte
	lw.State.SetField(tbl, "turn_absolute", func(l *lua.LState) int {
		angle := float64(l.ToNumber(1))
		turned := lw.Context.Player.TurnAbsolute(position.NewAngleDegree(angle), lw.ViewAngle)
		lw.Context.Player.UseMana()
		l.Push(lua.LNumber(turned))
	})
	lw.State.SetField(tbl, "go_ahead", func(l *lua.LState) int {
		distance := float64(l.ToNumber(1))
		walked := lw.Context.Player.GoAhead(distance)
		l.Push(lua.LNumber(walked))
	})
	lw.State.SetField(tbl, "get_mana", func(l *lua.LState) int {
		mana := lw.Context.Player.Mana
		l.Push(lua.LNumberBit(mana))
		return 1
	})
	lw.State.SetGlobal("player", tbl)
}
