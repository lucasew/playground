package blua

import (
	"github.com/lucasew/luaboard/board"
	"github.com/lucasew/luaboard/board/position"
	"github.com/yuin/gopher-lua"
	"log"
)

type LuaWrapper struct {
	State     *lua.LState
	Context   *board.PlayerContext
	ViewAngle position.Angle
	Functions map[string]map[string]lua.LGFunction
}

func (lw *LuaWrapper) Close() {
	log.Printf("Fechando contexto lua...\n")
	lw.State.Close()
}

func WrapContext(ctx *board.PlayerContext, viewAngle position.Angle) *LuaWrapper {
	log.Printf("Empacotando player context...\n")
	lw := &LuaWrapper{
		Context:   ctx,
		State:     lua.NewState(),
		ViewAngle: viewAngle,
	}
	LoadBattle(lw)
	LoadBoard(lw)
	LoadPlayer(lw)
	LoadPosition(lw)
	LoadUtils(lw)
	log.Printf("Adicionando funções ao Lua...\n")
	return lw
}
