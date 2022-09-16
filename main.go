package main

import (
	"fmt"
	"monkeyterm/src"
	"monkeyterm/src/core"
	//"monkeyterm/src/tui"
	"time"
)

func MakeTermLoop(g *src.Gutenberg) core.InputEventHandler {
  text := make([]byte,10) // assuming min 10 char
	return func(event core.InputEventInstance) {
		switch event.GetEventType() {
		case core.InputEvent_Rune:
			text = append(text, byte(event.GetEventValue()))
			g.Sections["Main"].Reprint(text)
		// default: 
  //     text = append(text, byte(event.GetEventValue()))
		// 	g.Sections["Main"].Reprint(text)
		} 
	}
	// return func (event *core.InputEventInstance) {
	// 	if event.Type == tui.Rune {
	// 		//r := rune(event.Value)
	// 		text = append(text, byte(event.Value))
	// 		g.Sections["Main"].Reprint(text)
 	//
	// 	  d := time.Since(start)
	// 		g.Sections["Header"].Reprint([]byte(fmt.Sprintf("Time: %d mils", d.Round(time.Millisecond))))
	// 	}
	// }

}

// func MakeTermLoop(g *src.Gutenberg) (func (event *tui.TermEvent)){
// 	text := make([]byte,10) // assuming min 10 char
// 	start := time.Now()
// 	return func (event *tui.TermEvent) {
// 		if event.Type == tui.Rune {
// 			//r := rune(event.Value)
// 			text = append(text, byte(event.Value))
// 			g.Sections["Main"].Reprint(text)
//
// 		  d := time.Since(start)
// 			g.Sections["Header"].Reprint([]byte(fmt.Sprintf("Time: %d mils", d.Round(time.Millisecond))))
// 		}
// 	}
	// return func (event *tui.TermEvent) {
	// 	if event.Type == tui.Rune {
	// 		r := rune(event.Value)
	// 		switch r {
	// 		case 'c':
	// 			t.GoToPageStart()
	// 		case 'e':
	// 			t.WriteRune('\n')
	// 		case 'q':
	// 			t.ClearToTopOfPage()
	// 		default:
	// 			t.OutputByte(byte(event.Value))
	// 		}
	// 	}
	// }
// }


func MakeHeaderHandler(g *src.Gutenberg) core.PollEventHandler{
	start := time.Now()
	return func(event core.PollEventInstance) {
		d := time.Since(start)
		g.Sections["Header"].Reprint([]byte(fmt.Sprintf("Time: %d mils", d.Round(time.Millisecond))))
	}
}

// func HeaderLoop(g *src.Gutenberg, running *bool) {	
// 	//for true {
// 		d := time.Since(start)
// 		time.Since(start)
// 		g.Sections["Header"].Reprint([]byte(fmt.Sprintf("Time: %d mils", d.Round(time.Millisecond))))
// 		time.Sleep(time.Duration(1000 * time.Millisecond))
// 	//}
// }

func main() {

	g := src.NewGutenberg()	
	// header section

	headerSection := src.NewSection(g, 1, 1, &src.SectionStyle{ContentAlign: src.Center})

	g.AddSection("Header", headerSection)

	mainSection := src.NewSection(g, 2, 2 , &src.SectionStyle{ContentAlign: src.Center})

	g.AddSection("Main", mainSection)

	g.RegisterPollHandler(MakeHeaderHandler(g))
	g.RegisterInputHandler(MakeTermLoop(g))

	//g.RegisterInputHandler(MakeTermLoop(g))
	g.MoveToSection("Main")
	g.Start()
	g.End()

}

