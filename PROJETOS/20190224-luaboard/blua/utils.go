package blua

import (
    "github.com/yuin/gopher-lua"
    "time"
    "math/rand"
    "math"
)

var mod_utils = func (lw *LuaWrapper) []LuaInitScript {
    return []LuaInitScript{
        RegisterFunction("delay", lw.LDelay),
        RegisterFunction("rand", lw.LRand),
        RegisterFunction("round", lw.LRound),
    }
}

func (lw *LuaWrapper) LRound(l *lua.LState) int {
    flt := float64(l.ToNumber(1))
    rounded := math.Round(flt)
    l.Push(lua.LNumber(rounded))
    return 1
}

func (lw *LuaWrapper) LDelay(l *lua.LState) int {
    msecs := l.ToInt(1)
    time.Sleep(time.Duration(msecs) * time.Millisecond)
    return 0
}


func (lw *LuaWrapper) LRand(l *lua.LState) int {
    max := float64(l.ToNumber(1))
    res := rand.Float64()*max
    l.Push(lua.LNumber(res))
    return 1
}
