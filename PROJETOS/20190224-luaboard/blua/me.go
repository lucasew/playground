package blua

import (
    "github.com/yuin/gopher-lua"
    "github.com/lucasew/luaboard/board/position"
)

var mod_me = func (lw *LuaWrapper) []LuaInitScript{
    return []LuaInitScript{
        RegisterFunction("me_getDistanceToPoint", lw.LGetDistanceToPoint),
        RegisterFunction("me_getDegreeToPoint", lw.LGetDegreeToPoint),
        RegisterFunction("me_getDegreeTowardsPoint", lw.LMeGetDegreeTowardsPoint),
        RegisterFunction("me_getMana", lw.LMeGetMana),
        RegisterFunction("me_getSpeed", lw.LMeGetSpeed),
        RegisterFunction("me_getHeading", lw.LMeGetHeading),
        RegisterFunction("me_getAtk", lw.LMeGetAtk),
        RegisterFunction("me_getDef", lw.LMeGetDef),
        RegisterFunction("me_getPos", lw.LMeGetPos),
        RegisterScriptString(`
            me = {
                get_distance_to_point = me_getDistanceToPoint,
                get_degree_to_point = me_getDegreeToPoint,
                get_degree_towards_point = me_getDegreeTowardsPoint,
                get_mana = me_getMana,
                get_speed = me_getSpeed,
                get_heading = me_getHeading,
                get_atk = me_getAtk,
                get_def = me_getDef,
                get_pos = me_getPos
            }
        `)}
}



func (lw *LuaWrapper) LGetDistanceToPoint(l *lua.LState) int {
    target := position.Position{
        X: float64(l.ToNumber(1)),
        Y: float64(l.ToNumber(2)),
    }
    res := lw.Context.Player.Position.GetDistance(target)
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LGetDegreeToPoint(l *lua.LState) int {
    x := float64(l.ToNumber(1))
    y := float64(l.ToNumber(2))
    res := lw.Context.GetDegreeTowardsPoint(x, y).Degree()
    l.Push(lua.LNumber(res))
    return 1
}


func (lw *LuaWrapper) LMeGetDegreeTowardsPoint(l *lua.LState) int {
    x := float64(l.ToNumber(1))
    y := float64(l.ToNumber(2))
    res := lw.Context.GetDegreeTowardsPoint(x, y).Degree()
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LMeGetMana(l *lua.LState) int {
    res := lw.Context.Player.Mana
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LMeGetSpeed(l *lua.LState) int {
    res := lw.Context.Player.Speed
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LMeGetHeading(l *lua.LState) int {
    res := lw.Context.Player.Position.GetHeading().Degree()
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LMeGetAtk(l *lua.LState) int {
    res := lw.Context.Player.AttackPoints
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LMeGetDef(l *lua.LState) int {
    res := lw.Context.Player.DefensePoints
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LMeGetPos(l *lua.LState) int {
    ret := lw.Context.Player.Position
    l.Push(lua.LNumber(ret.X))
    l.Push(lua.LNumber(ret.Y))
    return 2
}
