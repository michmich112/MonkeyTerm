package engine

import (
	"errors"
	"monkeyterm/src/core"
	"monkeyterm/src/tui"
	utils "monkeyterm/src/utils"
	"strings"
	"sync"

	wc "github.com/mattn/go-runewidth"
	"golang.org/x/exp/constraints"
)

type TermStyle struct {
  Bg string // Background Color
  Fg string // Foreground Color
}

type RenderStyle struct {
  NoInput TermStyle
  CorrectInput TermStyle
  WrongInput TermStyle
}

func DefaultRenderStyle() *RenderStyle {
  return &RenderStyle{
    NoInput: TermStyle{
      Bg: tui.GetTermColor(utils.BgDefault),
      Fg: tui.Get256FgTermColor(utils.Grey256),
    },
    CorrectInput: TermStyle{
      Bg: tui.GetTermColor(utils.BgDefault),
      Fg: tui.GetTermColor(utils.FgDefault),
    },
    WrongInput: TermStyle{
      Bg: tui.GetTermColor(utils.BgDefault),
      Fg: tui.GetTermColor(utils.FgRed),
    },
  }
}

type Integer constraints.Integer

type Dimension2D[T Integer] struct {
  Width T
  Height T
}

type MonkeyEngine[T Integer] struct {
  L sync.Mutex
  tokenizer Tokenizer
  inputBuffer []byte
  textBuffer [][]byte // text ordered in lines
  curLine int
  renderBuffer [][]byte
  curRenderLine int
  cursorLoc uint // location of the cursor, (where we are on the current line)
  canvas Dimension2D[T]
  lines uint
  eof bool // defaults to false (bool zero value)
}

func NewMonkeyEngine[T Integer](canvasWidth T, canvasHeight T) (*MonkeyEngine[T], error) {
  if canvasWidth <= 0 || canvasHeight <= 0 {
    return new(MonkeyEngine[T]), errors.New("Invalid canvas parameters")
  }

  var curLine int
  if canvasHeight > 1 {
    curLine = 1
  }

  return &MonkeyEngine[T]{
    canvas: Dimension2D[T]{
      Width: canvasWidth,
      Height: canvasHeight,
    },
    curRenderLine: curLine,
  }, nil
}

func (me *MonkeyEngine[T]) SetTokenizer(tokenizer Tokenizer) {
  me.tokenizer = tokenizer
}


func (me *MonkeyEngine[T]) init() {
  me.L.Lock() // good practice
  // Set the number of lines to render with a minimum of 1
  if t := me.canvas.Height - 1; t > 0 {
    me.lines = uint(t)
  } else {
    me.lines = uint(0)
  }

  // make the first like empty if there is more than one
  if me.lines > 1 {
    me.textBuffer = append(me.textBuffer, []byte(""))
  }
  
  // Generate the first renderable tokens
  token := ""
  for uint(len(me.textBuffer)) < me.lines{
    
    nl, t, eof := me.getNewLine([]byte(token))
    token = t // goply wasn't having it
    // make sure to append the new line
    me.textBuffer = append(me.textBuffer, nl)
    // if end of file we have nothing else to render
    if eof{ 
      break
    }
  }
  me.cursorLoc = 0 // not necessary but if user had changed it, idk.
  me.L.Unlock()
}



func (me *MonkeyEngine[T]) getNewLine(bufferedToken []byte) (line []byte, token string, eof bool) {
    lineBuf := make([]byte, 0, int(me.canvas.Width))
    // if there was a token that couldn't fit previously, we make it fit now
    if len(bufferedToken) != 0 {
      lineBuf = append(lineBuf, []byte(bufferedToken)...)
    }

    for T(wc.DefaultCondition.StringWidth(string(lineBuf))) < me.canvas.Width {
      token, eof := me.tokenizer.Next()
      if eof || T(wc.DefaultCondition.StringWidth(string(lineBuf)+token+" ")) > me.canvas.Width {
        break // token will have been set if not eof
      } else {
        lineBuf = append(lineBuf, []byte(token+" ")...)
      }
    } 
    return lineBuf, token, eof
}

// Refreshes the render bytes based on the input
func (me *MonkeyEngine[T]) refreshRender() {
  tokenizedInput := strings.Split(string(me.inputBuffer), " ")
  tokenizedRender := strings.Split(string(me.textBuffer[me.curLine])," ")
  // we only need to look at the last ones normally


}


func (me *MonkeyEngine[T]) handleNewRune(r rune) {
  me.L.Lock()
  me.inputBuffer = append(me.inputBuffer, byte(r))


}

func (me *MonkeyEngine[T]) GetInputHandler() core.InputEventHandler{

  return func(event core.InputEventInstance) {
    switch event.GetEventType() {
    case core.InputEvent_Rune:


    }
    me.L.Lock()
    me.L.Unlock() // make sure to unlock
  }

}

func (me *MonkeyEngine[T]) GetRenderText() []byte {
  return []byte{}
}

