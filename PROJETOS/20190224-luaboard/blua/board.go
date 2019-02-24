package blua

import (
    "github.com/yuin/gopher-lua"
)

var mod_board = func (lw *LuaWrapper) []LuaInitScript{
    return []LuaInitScript{
        RegisterFunction("board_get_size_x", lw.LBoardGetSizeX),
        RegisterFunction("board_get_size_y", lw.LBoardGetSizeY),
        RegisterScriptString(`
        function board_get_size()
            return board_get_size_x(), board_get_size_y()
        end
        `),
        RegisterFunction("board_count_enemies", lw.LBoardEnemyCount),
        RegisterScriptString(`
        board = {
            get_size_x = board_get_size_x,
            get_size_y =  board_get_size_y,
            get_size = board_get_size,
            count_enemies = board_count_enemies
        }
    `)}
}

func (lw *LuaWrapper) LBoardGetSizeX(l *lua.LState) int {
    res := lw.Context.Board.SizeX
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LBoardGetSizeY(l *lua.LState) int {
    res := lw.Context.Board.SizeY
    l.Push(lua.LNumber(res))
    return 1
}

func (lw *LuaWrapper) LBoardEnemyCount(l *lua.LState) int {
    res := len(lw.Context.Board.Players) - 1
    l.Push(lua.LNumber(res))
    return 1
}
