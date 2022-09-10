package main

import (
	"fmt"
	"monkeyterm/src/tui"
)

func main() {
	t, err := tui.NewTerm()
	if err != nil {
		panic(err)
	}

	defer t.Restore()

	t.OutputBytes([]byte("Hello World\n"))
	t.OutputBytes([]byte(fmt.Sprintf("x:%d, y:%d\n",t.InitialX(), t.InitalY())))
	x,y, _ := t.GetCursorPosition()
	t.OutputBytes([]byte(fmt.Sprintf("x:%d, y:%d\n",x,y)))
	t.WriteOutBuf()

	// Blocking Operation
	t.Start()
}

