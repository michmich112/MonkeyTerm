package src

import (
	"errors"
	"fmt"
	"monkeyterm/src/tui"
	"monkeyterm/src/core"
	"strings"
	"time"
	"sync"

	wc "github.com/mattn/go-runewidth"
)

// Gutenberg is the terminal printing service
type Gutenberg struct {
  O *sync.Mutex // output mutex that needs to be held for output operations
  Term *tui.Term
  Delay time.Duration // refresh delay ~ unsure if wil use this but we'll see
  Sections map[string]GutenbergSection // {sectionName: section}

	// Event Busses
	InputEventBus core.EventBus[core.InputEventValueType]
	PollEventBus core.EventBus[core.PollEventValueType]


  //Handler func (event *tui.TermEvent)
  InputHandlers []core.InputEventHandler 
  PollHandlers []core.PollEventHandler

  // events wg
  EventWg sync.WaitGroup
  
  //orderedSections []Section // sections ordered from top to bottom // will i need this? dunno
}

type GutenbergEvent struct {
	Trigger GutenbergEventTriggers
	T time.Time
}

type GutenbergEventTriggers = int

func NewGutenberg() (*Gutenberg) {
	t, err := tui.NewTerm([]int{tui.CtrlC})
	if err != nil {
		panic(err)
	}

	// t.OutputBytes([]byte("Hello World\n"))
	// t.OutputBytes([]byte(fmt.Sprintf("x:%d, y:%d\n",t.InitialX(), t.InitalY())))
	// x,y, _ := t.GetCursorPosition()
	// t.OutputBytes([]byte(fmt.Sprintf("x:%d, y:%d\n",x,y)))
	// t.WriteOutBuf()

	// Blocking Operation
	//t.Start(MakeTermLoop(t))

	return NewGutenbergFromTerm(t)
}

func NewGutenbergFromTerm(term *tui.Term) (*Gutenberg) {
	return &Gutenberg{
		O: &sync.Mutex{},
		Term: term,
		Delay: time.Duration(10000),
		Sections: make(map[string]GutenbergSection),
		// sectionHandlerWg: sync.WaitGroup{},
		InputEventBus: NewInputEventBus(),
		PollEventBus: NewPollEventBus(),
		InputHandlers: make([]core.InputEventHandler, 0),
		PollHandlers: make([]core.PollEventHandler, 0),

		EventWg: sync.WaitGroup{},
	}
}

func (g *Gutenberg) RegisterInputHandler(handler core.InputEventHandler) {
	g.InputHandlers = append(g.InputHandlers, handler)
}

func (g *Gutenberg) RegisterPollHandler(handler core.PollEventHandler) {
	g.PollHandlers = append(g.PollHandlers, handler)
}

// func (g *Gutenberg) RegisterHandler(handler func(event *tui.TermEvent)) {
// 	g.Handler = handler
// }

func (g *Gutenberg) SetDelay(delay time.Duration) {
	g.Delay = delay
}

func (g *Gutenberg) HasSection(name string) bool {
	_, exists := g.Sections[name]
	return exists
}

func (g *Gutenberg) AddSection(name string, section GutenbergSection) (error) {
	if g.HasSection(name) {
		//log.Warnf("Section with name %s already exists.", name)
		return errors.New(fmt.Sprintf("Section with name %s already exists. Remove or overwrite it", name))
	}
	g.Sections[name] = section
	return nil
}

func (g *Gutenberg) RefreshSection(name string) {}

func (g *Gutenberg) MoveToSection(name string) {
	if g.HasSection(name) {
		x,y := g.Sections[name].GetSectionOrigin()
		g.Term.GotoRelativePosition(x,y)
	}
}


// func (g *Gutenberg) Start() {
// 	SyncSectionHandlers := make([]GutenbergSectionHandler,0)
// 	for _, section := range(g.Sections) {
// 		if handler, err := section.GetHandler(); err != nil {
// 			if handler.IsAsync() {
// 				g.sectionHandlerWg.Add(1)
// 				go handler.Start()
// 			} else {
//
// 			}
// 		}
// 	}
// 	g.Term.Start(g.Handler)
// }

func (g *Gutenberg) termHandler(term *tui.Term, input byte, err error) {
	e := core.InputEvent{T: time.Now(), Value: int(input), Type: core.InputEvent_Rune}

	if err != nil {
		e.Type = core.InputEvent_ReadError
		defer g.InputEventBus.Close() // Close on error
	} else if term.IsEscapeCode(input) {
		e.Type = core.InputEvent_EscSequence
		defer g.InputEventBus.Close() // Close on esc seq
	} else	if input >= 0 && input < 32 || input == 127 {
		e.Type = core.InputEvent_CtrlChar
	}

	g.InputEventBus.Send(e)
}

func (g *Gutenberg) Start() {
	//g.EventWg.Add(3)	
	// input event handler goroutine
	go func() {
		c := g.InputEventBus.Receive()
		for inputEvent := range c {
			for _, handler := range g.InputHandlers {
				handler(inputEvent)
			}
		}
		// Make sure to close the bus channel
		//g.EventWg.Done()
	}()

	// poll event handler
	go func() {
		c := g.PollEventBus.Receive()

		for pollEvent := range c {
			for _, handler := range g.PollHandlers {
				handler(pollEvent)
			}
		}
		// Make sure to close the bus channel to reach here
		//g.EventWg.Done()
	}()

	// Polling GoRoutine
	//polling := true
	//go func() {
		//for polling {
			//time.Sleep(g.Delay)
			//g.PollEventBus.Send(core.PollEvent{T: time.Now()})
		//}
		//g.EventWg.Done()
	//}()


	g.Term.Start(g.termHandler)
	//polling = false
	//g.EventWg.Wait()
}

func (g *Gutenberg) End() {
	g.Term.Restore()
}

type GutenbergSection interface {
  // Reprints the entire section, not just the changes
  // Note: this is a time expensive op
  Reprint(text []byte) error

	// Wrapper for Reprint with fmt Sprintf for variables 
	// TODO - see if this is needed/useful
  //Reprintf(text string , argv ...interface{}) error

  // Reprints the changes
  Update(text []byte) error

  // Gets the relative cursor position for the section origin
  GetSectionOrigin() (x, y int)

  // Registers a handler
  //RegisterHandler(handler GutenbergSectionHandler)

  // Retrieve the handler
  //GetHandler() (GutenbergSectionHandler, error)

  // Called for clean clearing of the section
  CleanStop()

	// Wrapper for update using fmt.Sprintf to format the string
	// TODO - see if this is needed/useful
  // Updatef(text string , argv ...interface{}) error
}


type GutenbergSectionHandler interface {
	// is the handler sync (false) or async (true)
	IsAsync() bool

	// Start the Section Handler
	Start() error 

	// Stop the Section Handler
	Stop() error
}

type SectionStyle struct {
	//Wrap bool // default: False // not implemented
	ContentAlign SectionContentAlignment // default: Left
	// Padding
	// bg color?
	// fg color?
}

// A section is an area of printable space (defined in # of lines of the term)
type Section struct {
  gutenberg *Gutenberg
  Start int
  Size int
  Contents []byte // contents currently printed on the screen
  LastUpdate time.Time
  // Ui params
  Style *SectionStyle
  // Handler
  Handled bool
  Handler GutenbergSectionHandler 
}

type SectionContentAlignment int

const (
	Left SectionContentAlignment = iota // default
	Right
	Center
	//SpaceBetween 
	//SpaceEvenly
)

// Creates a New Sync GutenbergSection
// Does not clear the Section of whatever could already be on there
func NewSection(gutenberg *Gutenberg,start int, size int, style *SectionStyle) *Section {
	return &Section{
		gutenberg: gutenberg,
		Start: start,
		Size: size,
		LastUpdate: time.Now(),
		Style: style,
	}
}

// Reprints the entire Section, not just the changes
// This is an expensive op
func (s *Section) Reprint(text []byte) error {
	// Compute
	width, err := s.gutenberg.Term.GetWidth()
	if err != nil {
		// log.warn(unable to estimate term width, assuming input is proper width)
		width = wc.StringWidth(string(text))
	}
	lines := strings.Split(string(text), "\n") // seperate by newline - might have to check for \r
	for i,l := range(lines){
		switch s.Style.ContentAlign {
		case Left:
			lines[i] = leftAlignText(string(l), width )
		case Right:
			lines[i] = rightAlignText(string(l), width)
		case Center:
			lines[i] = centerAlignText(string(l), width)
		}
	}
	out := strings.Join(lines, "\n\r")
	// Print
	//s.gutenberg.O.TryLock()
	s.gutenberg.O.Lock() // Obtain lock to print
	tui.NewTermTx(s.gutenberg.Term).
	AddAction(func(t *tui.Term) {t.GotoRelativePosition(s.Start, 0)}).
	AddAction(func(t *tui.Term) {t.UnsafeOutputBytes([]byte(out))}).
	Execute()


	//x,y, err := s.gutenberg.Term.GetCursorPosition()
	//if err != nil {
		//panic("tset")
		//s.gutenberg.O.Unlock()
	//}

	// s.gutenberg.Term.GotoRelativePosition(s.Start,0)
	// s.gutenberg.Term.OutputBytes([]byte(out))
	// s.gutenberg.Term.GotoPosition(x,y)
	s.gutenberg.O.Unlock() // Release lock
	return nil
}

// Reprints only the updated sections of
func (s *Section) Update(text []byte) error {

	return nil
}

// returns col, row
func (s *Section) GetSectionOrigin() (x,y int) {
	return s.Start,0
}

// func (s *Section) RegisterHandler(handler GutenbergSectionHandler) {
// 	if !s.Handled {
// 		// Section is not set as a handled section
// 		s.Handled = true
// 	}
// 	s.Handler = handler // assign handler
// }
//
func (s *Section) CleanStop() {
	// Clean the canvas? maybe
	s.Handler.Stop() // stop the handler to possibly release the waitgroup
}

// The async implementation of a section
type AsyncSection struct {
  Start int
  Size int
  Contents []byte
  LastUpdate time.Time
  updateBuffer []byte
}

func (s *AsyncSection) Reprint() {

}

func (s *AsyncSection) Update() {

}

// Text Format Utilities

// Aligns the text to the center and fills the additional area with spaces
// Will only work for a single line
func centerAlignText(text string, width int) (string) {
	tw := wc.StringWidth(text)
	if tw > width { // cut off text
		return text[:width]
	}

	hw := (width-tw)>>1
	s := strings.Repeat(" ",hw)	
	if (hw<<1 + tw) < width { // add extra space since it was an odd number
		return fmt.Sprintf("%s%s%s ",s,text,s)
	}
	return fmt.Sprintf("%s%s%s",s,text,s)
}

// Aligns the text right and fills the additional area with spaces
// Will only work for a single line
func rightAlignText(text string, width int) (string) {
	tw := wc.StringWidth(text)
	if tw > width {
		return text[tw-width:]
	}
	return fmt.Sprintf("%s%s", strings.Repeat(" ", width-tw),text)
}

// Aligns the text left and fills the additional area with spaces
// Will only work for a single line
func leftAlignText(text string, width int) (string) {
	tw := wc.StringWidth(text)
	if tw > width {
		return text[:tw-width]
	}
	return fmt.Sprintf("%s%s", text, strings.Repeat(" ", width-tw))
}


