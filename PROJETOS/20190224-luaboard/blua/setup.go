package blua

import (
    "github.com/yuin/gopher-lua"
    "log"
)

type LuaInitScript func(*LuaWrapper)error

func RegisterFunction(name string, fn func(*lua.LState) int) LuaInitScript {
    log.Printf("INIT: Registrando função: %v\n", name)
    return func(lw *LuaWrapper)error {
            lw.State.SetGlobal(name, lw.State.NewFunction(fn))
            return nil
    }
}

func RegisterScriptString(script string) LuaInitScript {
    log.Printf("INIT: Registrando script: \n%s\n", script)
    return func(lw *LuaWrapper)error {
            return lw.State.DoString(script)
    }
}

func Apply(scripts []LuaInitScript , ctx *LuaWrapper) error {
    log.Printf("INIT: Aplicando scripts ao contexto do jogador...\n")
    for _, v := range(scripts) {
        err := v(ctx)
        if err != nil {
            return err
        }
    }
    return nil
}
