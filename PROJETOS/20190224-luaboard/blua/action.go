package blua

import (
    "github.com/yuin/gopher-lua"
    "github.com/lucasew/luaboard/board/position"
)

var mod_action = func (lw *LuaWrapper) []LuaInitScript {
    return []LuaInitScript{
        RegisterFunction("action_Turn", lw.LActionTurn),
        RegisterFunction("action_GoAhead", lw.LActionGoAhead),
        RegisterScriptString(`
        action = {
            turn = action_Turn,
            go_ahead = action_GoAhead
        }
        `)}
}

func (lw *LuaWrapper) LActionTurn(l *lua.LState) int {
    degree := l.ToNumber(1)
    angle := position.NewAngleDegree(float64(degree))
    target := lw.Context.Player.Position.DOTurn(angle, lw.ViewAngle)
    l.Push(lua.LNumber(target.Degree()))
    return 1
}

func (lw *LuaWrapper) LActionGoAhead(l *lua.LState) int {
    distance := l.ToNumber(1)
    destination := lw.Context.GoAhead(float64(distance))
    l.Push(lua.LNumber(destination.X))
    l.Push(lua.LNumber(destination.Y))
    return 2
}
