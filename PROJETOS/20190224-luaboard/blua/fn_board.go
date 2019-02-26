package blua

import (
	"github.com/lucasew/luaboard/board/position"
	"github.com/yuin/gopher-lua"
)

func LoadBoard(lw *LuaWrapper) {
	funcs := map[string]lua.LGFunction{}
	funcs["get_view_angle"] = func(l *lua.LState) int {
		angle := lw.ViewAngle.Degree()
		l.Push(lua.LNumber(angle))
		return 1
	}
	funcs["get_size_x"] = func(l *lua.LState) int {
		sizex := lw.Context.Board.SizeX
		l.Push(lua.LNumber(sizex))
		return 1
	}
	funcs["get_size_y"] = func(l *lua.LState) int {
		sizey := lw.Context.Board.SizeY
		l.Push(lua.LNumber(sizey))
		return 1
	}
	funcs["get_size"] = func(l *lua.LState) int {
		sizex := lw.Context.Board.SizeX
		sizey := lw.Context.Board.SizeY
		l.Push(lua.LNumber(sizex))
		l.Push(lua.LNumber(sizey))
		return 2
	}
	// TODO: Implementar filtro para apenas contar players vivos
	funcs["count_enemies"] = func(l *lua.LState) int {
		count := len(lw.Context.Board.Players) - 1
		l.Push(lua.LNumber(count))
		lw.Context.Player.UseMana()
		return 1
	}
	tbl := lw.State.NewTable()
	lw.State.SetFuncs(tbl, funcs)
	lw.State.SetGlobal("board", tbl)
}
