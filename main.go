package main

import (
	"fmt"
	"monkeyterm/src/tui"
)

func MakeTermLoop(t *tui.Term) (func (event *tui.TermEvent)){
	return func (event *tui.TermEvent) {
		if event.Type == tui.Rune {
			r := rune(event.Value)
			switch r {
			case 'c':
				t.GoToPageStart()
			case 'e':
				t.WriteRune('\n')
			case 'q':
				t.ClearToTopOfPage()
			default:
				t.OutputByte(byte(event.Value))
			}
		}
	}
}

func main() {
	t, err := tui.NewTerm([]int{tui.CtrlC})
	if err != nil {
		panic(err)
	}

	defer t.Restore()

	t.OutputBytes([]byte("Hello World\n"))
	t.OutputBytes([]byte(fmt.Sprintf("x:%d, y:%d\n",t.InitialX(), t.InitalY())))
	x,y, _ := t.GetCursorPosition()
	t.OutputBytes([]byte(fmt.Sprintf("x:%d, y:%d\n",x,y)))

	// Blocking Operation
	t.Start(MakeTermLoop(t))
}

