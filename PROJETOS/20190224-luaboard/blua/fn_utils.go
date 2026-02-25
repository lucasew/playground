package blua

import (
	"github.com/yuin/gopher-lua"
	"math"
	"math/rand"
	"time"
)

func LoadUtils(lw *LuaWrapper) {
	funcs := map[string]lua.LGFunction{}
	funcs["round"] = func(l *lua.LState) int {
		flt := float64(l.ToNumber(1))
		rounded := math.Round(flt)
		l.Push(lua.LNumber(rounded))
		return 1
	}
	funcs["delay"] = func(l *lua.LState) int {
		msecs := int(l.ToInt(1))
		time.Sleep(time.Duration(msecs) * time.Millisecond)
		return 0
	}
	funcs["rand"] = func(l *lua.LState) int {
		max := float64(l.ToNumber(1))
		res := rand.Float64() * max
		l.Push(lua.LNumber(res))
		return 1
	}
	tbl := lw.State.NewTable()
	lw.State.SetFuncs(tbl, funcs)
	lw.State.SetGlobal("util", tbl)
}
