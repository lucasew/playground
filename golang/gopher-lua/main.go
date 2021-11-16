package main

import (
	"fmt"
	"time"

	"github.com/yuin/gopher-lua"
    // luajson "layeh.com/gopher-json"
)

func main() {
    L := lua.NewState(lua.Options{
        SkipOpenLibs: true,
    })
    // luajson.Preload(L)
    L.SetGlobal("sleep", L.NewFunction(func(L *lua.LState) int {
        ms := L.CheckInt(1)
        time.Sleep(time.Millisecond*time.Duration(ms))
        return 0
    }))
    L.Push(L.NewFunction(lua.OpenBase))
    L.Push(lua.LString(lua.BaseLibName))
    L.Call(1, 0)
    L.Push(L.NewFunction(lua.OpenString))
    L.Push(lua.LString(lua.StringLibName))
    L.Call(1, 0)
    L.Push(L.NewFunction(lua.OpenIo))
    L.Push(lua.LString(lua.IoLibName))
    L.Call(1, 0)
    L.Push(L.NewFunction(lua.OpenMath))
    L.Push(lua.LString(lua.MathLibName))
    L.Call(1, 0)
    defer L.Close()
    if err := L.DoString(`
    -- json = require('json')
    metatable = {
        __gc = function(self)
            print("Destruindo...")
        end
    }
    new = {}
    print("Eoq")
    new = setmetatable(new, metatable)
    sleep(100)
    collectgarbage()
     keys = function (t) do
        for k, v in pairs(t) do
            print(k .. ": " .. tostring(v))
        end
     end
    end
    -- keys(_G)

    `); err != nil {
        panic(err)
    }
    fn, err := L.LoadString(`return 2 + 2`)
    if err != nil {
        panic(err)
    }
    L.Push(fn)
    L.Call(0, 1)
    fmt.Println(L.CheckInt(1))
}
