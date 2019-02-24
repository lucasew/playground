package blua

import (
    "github.com/lucasew/luaboard/board"
    "github.com/lucasew/luaboard/board/position"
    "github.com/yuin/gopher-lua"
    "log"
)

type LuaWrapper struct {
    State *lua.LState;
    Context *board.PlayerContext;
    ViewAngle position.Angle;
}

func (lw *LuaWrapper) Close() {
    log.Printf("Fechando contexto lua...\n")
    lw.State.Close()
}

func WrapContext(ctx *board.PlayerContext, viewAngle position.Angle) *LuaWrapper {
    log.Printf("Empacotando player context...\n")
    lw := &LuaWrapper{
        Context: ctx,
        State: lua.NewState(),
        ViewAngle: viewAngle,
    }
    scripts := GetScripts(lw)
    log.Printf("Adicionando funções ao Lua...\n")
    for k, script := range(scripts) {
        log.Printf("Carregando script %d/%d", k + 1, len(scripts))
        err := script(lw)
        if err != nil {
            panic(err)
        }
    }
    return lw
}

