package blua

import (
	"github.com/lucasew/luaboard/board/position"
	"github.com/yuin/gopher-lua"
)

func LoadBattle(lw *LuaWrapper) {
	tbl := lw.State.NewTable()
	// TODO: Implementar isso no core
	lw.State.SetField(tbl, "is_enemy_near", func(l *lua.LState) int {
		distance := lw.Context.IsEnemyNear()
		if distance == 0 {
			return 0
		} else {
			l.Push(lua.LNumber(distance))
			return 1
		}
	})
	// TODO: Implementar isso no core tbm
	lw.State.SetField(tbl, "seen_enemy", func(l *lua.LState) int {
		isseenenemy := lw.Context.IsSeenEnemy()
		l.Push(lua.LBool(isseenenemy))
		return 1
	})
	lw.State.SetGlobal("battle", tbl)
}
