package blua

import (
	"github.com/lucasew/luaboard/board/position"
	"github.com/yuin/gopher-lua"
)

func LoadPlayer(lw *LuaWrapper) {
	funcs := map[string]lua.LGFunction{}
	funcs["get_x"] = func(l *lua.LState) int {
		res := lw.Context.Player.Position.X
		l.Push(lua.LNumber(res))
		return 1
	}
	funcs["get_y"] = func(l *lua.LState) int {
		res := lw.Context.Player.Position.Y
		l.Push(lua.LNumber(res))
		return 1
	}
	funcs["get_heading"] = func(l *lua.LState) int {
		res := lw.Context.Player.Position.Heading.Degree()
		l.Push(lua.LNumber(res))
		return 1
	}
	funcs["get_pos"] = func(l *lua.LState) int {
		resx := lw.Context.Player.Position.X
		l.Push(lua.LNumber(resx))
		resy := lw.Context.Player.Position.Y
		l.Push(lua.LNumber(resy))
		return 2
	}
	// TODO: Arrumar essa parte
	funcs["turn_relative"] = func(l *lua.LState) int {
		angle := float64(l.ToNumber(1))
		turned := lw.Context.Player.TurnRelative(position.NewAngleDegree(angle), lw.ViewAngle)
		lw.Context.Player.UseMana()
		l.Push(lua.LNumber(turned))
		return 1
	}
	// TODO: Arrumar essa parte
	funcs["turn_absolute"] = func(l *lua.LState) int {
		angle := float64(l.ToNumber(1))
		turned := lw.Context.Player.TurnAbsolute(position.NewAngleDegree(angle), lw.ViewAngle)
		lw.Context.Player.UseMana()
		l.Push(lua.LNumber(turned))
		return 1
	}
	funcs["go_ahead"] = func(l *lua.LState) int {
		distance := float64(l.ToNumber(1))
		walked := lw.Context.Player.GoAhead(distance)
		l.Push(lua.LNumber(walked))
		return 1
	}
	funcs["get_mana"] = func(l *lua.LState) int {
		mana := lw.Context.Player.Mana
		l.Push(lua.LNumber(mana))
		return 1
	}
	tbl := lw.State.NewTable()
	lw.State.SetFuncs(tbl, funcs)
	lw.State.SetGlobal("player", tbl)
}
