package blua

import (
	"github.com/yuin/gopher-lua"
	"math"
	"math/rand"
	"time"
)

func LoadUtils(lw *LuaWrapper) {
	lw.State.SetGlobal("round", func(l *lua.LState) int {
		flt := float64(l.ToNumber(1))
		rounded := math.Round(flt)
		l.Push(lua.LNumber(rounded))
		return 1
	})
	lw.State.SetGlobal("delay", func(l *lua.LState) int {
		msecs := l.ToInt(1)
		time.Sleep(time.Duration(msecs) * time.Millisecond)
		return 0
	})
	lw.State.SetGlobal("rand", func(l *lua.LState) int {
		max := float64(l.ToNumber(1))
		res := rand.Float64() * max
		l.Push(lua.LNumber(res))
		return 1
	})
}
