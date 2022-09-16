package tui

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	_ "time"

	"golang.org/x/exp/slices"
	"golang.org/x/term"
)



type Tty struct {
	in *os.File
	out *os.File
	fd int
}

type Term struct {
	L *sync.Mutex
	initialPosition [2]int
	tty *Tty
	isRaw bool
	saved *term.State
	terminal *term.Terminal	
	obL *sync.Mutex
	outBuf []byte
	escapeCodes []int
	started bool
	reqBuf bool
	escBuf []byte
}

func NewTerm(escapeCodes []int) (*Term, error){
	if len(escapeCodes) == 0 {
		return new(Term), errors.New("At least one escape code is required.")
	}

	tin := os.Stdin
	// tin := TtyIn()

	tty := &Tty{
		in: tin,
		out: os.Stderr,
		fd: int(tin.Fd()),
	}

	if !term.IsTerminal(tty.fd) {
		return new(Term), errors.New("Tty in not a terminal.")
	}

	saved, err := term.MakeRaw(tty.fd)
	if err != nil {
		return new(Term), err
	}

	nt := term.NewTerminal(tty.in, "")


	// change from blocking to nonblock
	syscall.SetNonblock(int(tin.Fd()), false)

	t := &Term{
		L: &sync.Mutex{},
		// initialPosition: [2]int{x,y},
		tty: tty,
		isRaw: true,
		saved: saved,
		terminal: nt,
		escapeCodes: escapeCodes,
	}

	x, y, err := t.getTerminalCursorPosition(tty.in)
	if err != nil {
		return new(Term), err
	}

	t.initialPosition = [2]int{x,y}


	return t,nil
}




// get the current x, y postion of the cursor for any tty (should be put in raw mode)
// the lock should be held when doing this operation
func (t *Term) getTerminalCursorPosition(tty *os.File) (x,y int, err error) {
	tty.Write([]byte("\x1B[6n"))

	var reader *bufio.Reader
	if t.started {
		r := bytes.NewReader(t.escBuf)
		reader = bufio.NewReader(r)
	} else {
		reader = bufio.NewReader(tty)
	}
	// text := make([]byte, 10)
	// _, err = reader.Read(text)
	text, err := reader.ReadSlice('R')

	if err != nil {
		return 0,0, err
	}
	fmt.Printf("Nst %+v", text)
	if slices.Contains(text, byte(';')) {
		re := regexp.MustCompile(`\d+;\d+`)
		pos := strings.Split(re.FindString(string(text)), ";")
		x, errx := strconv.Atoi(pos[0])
		y, erry := strconv.Atoi(pos[1])
		if errx != nil || erry != nil {
			return 0,0, errors.New("Error parsing pos value to int")
		}
		return x,y, nil
	}
	return 0, 0, errors.New("Invalid Return from ANSI Sequence")
}

func (t *Term) MakeCursorInvisible() {
	t.L.Lock()
	t.outputBytes([]byte("\x1B[?25l"))
	t.writeOutBuf()
	t.L.Unlock()
}

func (t *Term) MakeCursorVisible() {
	t.L.Lock()
	t.outputBytes([]byte("\x1B[?25h"))
	t.writeOutBuf()
	t.L.Unlock()
}

func (t *Term) InitialX() (int) {
	return t.initialPosition[0]
}

func (t *Term) InitalY() (int) {
	return t.initialPosition[1]
}

// Gets the available width of the terminal
func (t *Term) GetWidth() (int, error) {
	x,_,err :=  term.GetSize(t.tty.fd)
	if err != nil {
		return 0, err
	}
	return x, nil
}

// Gets the available inline heigh based on the starting location
func (t *Term) GetAvailableHeight() (int, error) {
	if y, err := t.GetHeight(); err != nil {
		return y, err
	} else {
		return y-t.InitalY(), nil
	}
}

func (t *Term) GetHeight() (int, error) {
	_, y, err := term.GetSize(t.tty.fd)
	if err != nil {
		return 0, err
	}
	return y, nil
}

// gets the current x,y position of the cursor
func (t *Term) GetCursorPosition() (x,y int, err error){
	t.L.Lock()
	x,y, err = t.getTerminalCursorPosition(t.tty.in)
	t.L.Unlock()
	return
}

// gets the curren x,y position of the cursor relative to the initial starting postition (only y)
func (t *Term) GetRelativeCursorPosition() (x,y int, err error) {
	x, y, err =  t.GetCursorPosition()
	if err == nil {
		y = y-t.InitalY()
	}
	return
}

// gets the current x,y position of the cursor
// Lock must be held
// !!unsafe
func (t *Term) getCursorPosition() (x, y int, err error) {
	x,y, err = t.getTerminalCursorPosition(t.tty.in)
	return
}

func (t *Term) PeekUserInput() ([]byte, error) {
	t.L.Lock()
	defer t.L.Unlock()
	reader := bufio.NewReader(t.tty.in)
	if reader.Buffered() > 0 {
		a, err:= reader.Peek(reader.Buffered())
		if err != nil {
			return []byte{}, err
		}
		return a, err
	}
  return []byte{}, nil
}

// Appends an ANSI Goto position code to the output buffer
func (t *Term) bufGotoPosition(x,y int) {
	t.outBuf = append(t.outBuf, []byte(fmt.Sprintf("\x1B[%d;%dH",x,y))...)
	//_, err := t.tty.in.Write([]byte(fmt.Sprintf("\x1B[%d;%dH",y,x)))
}

// Goes to the defined position on the terminal (not relative to starting point)
func (t *Term) GotoPosition(x, y int) {
	// maybe we should check if the values are outside the available canvas
	// t.L.Lock()
	t.bufGotoPosition(x, y)
	t.writeOutBuf() // write out all from the buffer
	// t.L.Unlock()
} 

// Goes to the desired position on the terminal relative to the starting point
func (t *Term) GotoRelativePosition(x, y int) {
	t.GotoPosition(x+t.InitialX(), y+t.InitalY())
}

// Appends an ANSI earase line code to the output buffer
func (t *Term) bufEraseEntireLine() {
	t.outBuf = append(t.outBuf, []byte("\x1B[2K")...)
}

// Appending an ANSI Cursor Down Character to the output buffer
func (t *Term) bufCursorDown() {
	t.outBuf = append(t.outBuf, []byte("\x1B[1B")...)
	//_, err := t.tty.in.Write([]byte("\x1B[1B"))
}

// adds byte to output buffer to be printed
// Lock must be held
// !!unsafe
func (t *Term) outputByte(byte byte) {
	t.outBuf = append(t.outBuf, byte)
}

// adds byte to output buffer to be printed
func (t *Term) OutputByte(byte byte) {
	t.L.Lock()
	t.outputByte(byte)
	t.L.Unlock()
}

// adds bytes to output buffer to be printed
// Lock must be held
// !!unsafe
func (t *Term) outputBytes(bytes []byte) {
	t.outBuf = append(t.outBuf, bytes...)
}

func (t *Term) UnsafeOutputBytes(bytes []byte) {
	t.outputBytes(bytes)
}

// adds bytes to output buffer to be printed
func (t *Term) OutputBytes(bytes []byte) {
	t.L.Lock()
	t.outputBytes(bytes)
	t.L.Unlock()
}

// adds rune to output buffer to be printed
// Lock must be held
// !!unsafe
func (t *Term) writeRune(r rune) {
	t.outBuf = append(t.outBuf, byte(r))
}

// adds rune to output buffer to be printed
func (t *Term) WriteRune(r rune) {
	t.L.Lock()
	t.writeRune(r)
	t.L.Unlock()
}


// Writes out the current contents of the output buffer until empty
func (t *Term) WriteOutBuf() (error) {
	t.L.Lock()
	err := t.writeOutBuf()
	t.L.Unlock()
	return err
}

// Writes out the current contents of the buffer
// Lock must be held before calling this
// !!unsafe
func (t *Term) writeOutBuf() (error) {
	l := len(t.outBuf)
	for l > 0 {
		ol, err := t.terminal.Write(t.outBuf)
		if err != nil {
			return err
		}
		if ol < l {
			// only remove the amount that was written out
			t.outBuf = t.outBuf[ol:]
			l = len(t.outBuf)
		} else {
			// keep the underlying array by sliceing to zero length
			t.outBuf = t.outBuf[:0]
			l = 0
		}
	}
	return nil
}

func (t *Term) GoToPageStart() (error) {
	t.L.Lock()
	t.bufGotoPosition(t.initialPosition[0], t.initialPosition[1])
	err := t.writeOutBuf()
	t.L.Unlock()
	return err
}

func (t *Term) ClearToTopOfPage() (error) {
	t.L.Lock()
	// write out buffer so that all pending contents are good to go
	// could technically be removed since we're going to clear it eitherway
	t.writeOutBuf()
	curx, cury, err := t.getCursorPosition()
	if err != nil {
		t.L.Unlock()
		return err
	}

	// fill output buffer with the action codes
	t.bufGotoPosition(t.initialPosition[0], t.initialPosition[1])
	for i := 0;i<=curx-t.initialPosition[0];i++ {
		t.bufEraseEntireLine()
		t.bufCursorDown()
	}
	t.bufGotoPosition(curx, cury)

	// write buffer
	t.writeOutBuf()	

	t.L.Unlock()
	return nil
}

// func (t *Term) WritePage([]byte) (error) {
// 	t.L.Lock()
//
// 	// Todo Implement this
//
// 	t.L.Unlock()
// 	return nil
// }


// restores the terminal to its pervious state
func (t *Term) Restore() {
	if !t.isRaw {
		panic("Unable to restore a non raw term")
	}
	t.isRaw = false
	term.Restore(t.tty.fd,t.saved)
}

func (t *Term) ClearCanvas(){

}

func (t *Term) ClearExit() {
	t.GoToPageStart()
}

func (t *Term) IsEscapeCode(input byte) bool {
	return slices.Contains(t.escapeCodes, int(input))
}

// func (t *Term) Start(actionFn func (event *TermEvent)) {
// 	wg := sync.WaitGroup{}
//
// 	go func () {
// 		reader := bufio.NewReader(t.tty.in)
// 		b := make([]byte,1)
// 		event := new(TermEvent)
// 		active := true
// 		for active {
// 			l, err := reader.Read(b)
// 			t.WriteRune('a')
// 			t.WriteOutBuf()
// 			if err != nil {
// 				active = false
// 				defer fmt.Printf("Read Error: %+v\n",err)
// 			}
// 			if l > 0 {
// 				if slices.Contains(t.escapeCodes, int(b[0])) {
// 					active = false	
//
// 				} else {
// 					event.UpdateFromRune(rune(b[0]))
// 					actionFn(event)
// 				}
// 				if len(t.outBuf) > 0 {
// 					t.WriteOutBuf()
// 				}
// 			}
// 		}
// 		t.ClearExit()
// 		wg.Done()
// 	}()
//
// 	wg.Add(1)
// 	wg.Wait()
// }

var CSI_END_CHAR = []rune{
	'R', // Return from cursor position request
}

// Reads the buffer and returns weather the escape code is complete
func (t *Term) HandleEscCode(buffer []byte) bool {
	// Manage CSI
	if buffer[0] == '\x9B' {
		if slices.Contains(CSI_END_CHAR, rune(buffer[len(buffer)-1])) {
			return true
		}
	}
	return false // this not good
}

func (t *Term) Start(handler func (t *Term, input byte, err error)) {
	wg := sync.WaitGroup{}

	go func () {
		reader := bufio.NewReader(t.tty.in)
		b := make([]byte,1)
		t.started = true
		// event := new(TermEvent)
		active := true

		escCodeBuffer := make([]byte,0)
		buffering := false
		for active {
			l, err := reader.Read(b)

		  //t.WriteRune('a')
			// t.WriteOutBuf()
			if err != nil {
				active = false
				fmt.Printf("Read Error: %+v\n",err)
				handler(t, 0, err) // handler with error
				panic("test")
			}

   		//
			// if (t.reqBuf) {
			// 	t.bufIn = append(t.bufIn, b[0])
			// }
   		//

			if l > 0 {
				if !buffering {
					if rune(b[0]) == '\x9B' {
						buffering = true
						escCodeBuffer = append(escCodeBuffer, b[0])
						continue // no op
					}
					if slices.Contains(t.escapeCodes, int(b[0])) {
						active = false	
					} 	
					handler(t, b[0], nil) // handler call

					if len(t.outBuf) > 0 {
						t.WriteOutBuf()
					}
				} else {
					escCodeBuffer = append(escCodeBuffer,b[0])
					if t.HandleEscCode(escCodeBuffer) { // this is the end of the esc sequence
						// Share Escape Code
						t.escBuf = escCodeBuffer
						escCodeBuffer = escCodeBuffer[:0]
						buffering = false
					}
				}
			}
		}
		t.ClearExit()
		wg.Done()
	}()

	wg.Add(1)
	wg.Wait()
}


func (t *Term) SaveCursorPosition() {
	t.OutputBytes([]byte("\x1B 7"))
}

func (t *Term) RestoreCursorPosition() {
	t.OutputBytes([]byte("\x1B 8"))
}

type TermTx struct {
	t *Term
	actions []func(t *Term)
}

func NewTermTx(t *Term) *TermTx{
	return &TermTx{
		actions: make([]func(t *Term), 0),
		t: t,
	}
}

//Term transaction builder
func (t *TermTx) AddAction(f func(t *Term)) *TermTx {
	t.actions = append(t.actions, f)
	return t
}

func (t *TermTx) Execute() {
	t.t.L.Lock()
	for _, f := range(t.actions) {
		f(t.t)
	}
	t.t.writeOutBuf() //write out all changes
	t.t.L.Unlock()
}


