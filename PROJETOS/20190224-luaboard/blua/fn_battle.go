package blua

import (
	"github.com/lucasew/luaboard/board/position"
	"github.com/yuin/gopher-lua"
)

func LoadBattle(lw *LuaWrapper) {
	funcs := map[string]lua.LGFunction{}
	// TODO: Implementar isso no core
	funcs["is_enemy_near"] = func(l *lua.LState) int {
		distance := lw.Context.IsEnemyNear()
		if distance == 0 {
			return 0
		} else {
			l.Push(lua.LNumber(distance))
			return 1
		}
	}
	// TODO: Implementar isso no core tbm
	funcs["seen_enemy"] = func(l *lua.LState) int {
		isseenenemy := lw.Context.IsSeenEnemy()
		l.Push(lua.LBool(isseenenemy))
		return 1
	}
	tbl := lw.State.NewTable()
	lw.State.SetFuncs(tbl, funcs)
	lw.State.SetGlobal("battle", tbl)
}
