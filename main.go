package main

import (
	"fmt"
	"monkeyterm/src"
	"monkeyterm/src/core"
	"time"
)

func MakeTermLoop(g *src.Gutenberg) core.InputEventHandler {
  text := make([]byte,10) // assuming min 10 char
	return func(event core.InputEventInstance) {
		switch event.GetEventType() {
		case core.InputEvent_Rune:
			text = append(text, byte(event.GetEventValue()))
			g.Sections["Main"].Reprint(text)
		} 
	}
}

func MakeHeaderHandler(g *src.Gutenberg) core.PollEventHandler{
	start := time.Now()
	return func(event core.PollEventInstance) {
		d := time.Since(start)
		g.Sections["Header"].Reprint([]byte(fmt.Sprintf("Time: %d mils", d.Round(time.Millisecond))))
	}
}

func main() {

	g := src.NewGutenberg()	
	// header section

	headerSection := src.NewSection(g, 1, 1, &src.SectionStyle{ContentAlign: src.Center})

	g.AddSection("Header", headerSection)

	mainSection := src.NewSection(g, 2, 2 , &src.SectionStyle{ContentAlign: src.Center})

	g.AddSection("Main", mainSection)

	footer := src.NewSection(g, 4,1, &src.SectionStyle{ContentAlign: src.Right})

	g.AddSection("Footer", footer)

	g.RegisterPollHandler(MakeHeaderHandler(g))
	g.RegisterInputHandler(MakeTermLoop(g))
	g.RegisterPollHandler(func(event core.PollEventInstance) {g.Sections["Footer"].Reprint([]byte(event.GetEventValue().GoString()))})

	//g.RegisterInputHandler(MakeTermLoop(g))
	g.MoveToSection("Main")
	g.Start()
	g.End()

}

