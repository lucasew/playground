package main

import (
	"bufio"
	"fmt"
	"github.com/lucasew/luaboard/blua"
	"github.com/lucasew/luaboard/board"
	"github.com/lucasew/luaboard/board/position"
	"os"
	"time"
)

var ViewAngle = position.NewAngleDegree(30)

func main() {
	brd := board.NewBoard(10, 10)
	player := brd.NewPlayer(10, 10, 3)
	brd.InitBoard()
	brd.StartManaTicker(time.Millisecond * 400)
	brd.StartSanitizerScheduler(time.Second * 1)
	control := blua.WrapContext(player, ViewAngle)
	defer control.Close()
	err := control.State.DoFile("./actions.lua")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("->>  ")
		cmd, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		err = control.State.DoString(cmd)
		fmt.Println(err)
	}
}
