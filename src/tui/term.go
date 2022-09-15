package tui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

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
}

func NewTerm(escapeCodes []int) (*Term, error){
	if len(escapeCodes) == 0 {
		return new(Term), errors.New("At least one escape code is required.")
	}

	tty := &Tty{
		in: os.Stdin,
		out: os.Stdout,
		fd: int(os.Stdin.Fd()),
	}

	if !term.IsTerminal(tty.fd) {
		return new(Term), errors.New("Tty in not a terminal.")
	}

	saved, err := term.MakeRaw(tty.fd)
	if err != nil {
		return new(Term), err
	}
	
	x, y, err := getTerminalCursorPosition(tty.in)
	if err != nil {
		return new(Term), err
	}

	nt := term.NewTerminal(tty.in, "")

	t := &Term{
		L: &sync.Mutex{},
		initialPosition: [2]int{x,y},
		tty: tty,
		isRaw: true,
		saved: saved,
		terminal: nt,
		escapeCodes: escapeCodes,
	}
	return t,nil
}

// get the current x, y postion of the cursor for any tty (should be put in raw mode)
func getTerminalCursorPosition(tty *os.File) (x,y int, err error) {
	tty.Write([]byte("\x1B[6n"))
	reader := bufio.NewReader(tty)
	text, err := reader.ReadSlice('R')
	if err != nil {
		return 0,0, err
	}
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

func (t *Term) InitialX() (int) {
	return t.initialPosition[0]
}

func (t *Term) InitalY() (int) {
	return t.initialPosition[1]
}

// gets the current x,y position of the cursor
func (t *Term) GetCursorPosition() (x,y int, err error){
	t.L.Lock()
	x,y, err = getTerminalCursorPosition(t.tty.in)
	t.L.Unlock()
	return
}

// gets the current x,y position of the cursor
// Lock must be held
// !!unsafe
func (t *Term) getCursorPosition() (x, y int, err error) {
	x,y, err = getTerminalCursorPosition(t.tty.in)
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


func (t *Term) Start(actionFn func (event *TermEvent)) {
	reader := bufio.NewReader(t.tty.in)
	b := make([]byte,1)
	event := new(TermEvent)
	active := true
	for active {
		l, err := reader.Read(b)
		if err != nil {
			active = false
			defer fmt.Printf("Read Error: %+v\n",err)
		}
		if l > 0 {
			if slices.Contains(t.escapeCodes, int(b[0])) {
				active = false	
			} else {
				event.UpdateFromRune(rune(b[0]))
				actionFn(event)
			}
			if len(t.outBuf) > 0 {
				t.WriteOutBuf()
			}
		}
	}
	t.ClearExit()
}

