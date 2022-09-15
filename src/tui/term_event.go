package tui

import (
  "time"
)

type TermEvent struct {
	Type TermEventType // term event type
	Value int // value
	T time.Time // timestamp of the event
}

func NewTermEvent(eventType TermEventType, value int) *TermEvent{
  return &TermEvent{
    Type: eventType,
    Value: value,
    T: time.Now(),
  }
}

// In Place Update
// updates the data of the event at the memory address
func (t *TermEvent) Update(eventType TermEventType, value int) {
  t.Type = eventType
  t.Value = value
  t.T = time.Now()
}

// Updates the current event from the rune
func (t *TermEvent) UpdateFromRune(value rune) {
  if value >  31 && value < 127 {
    t.Update(Rune, int(value))
  } else {
    t.Update(CtrlChar, int(value))
  }
}

type TermEventType int

const (
  Rune TermEventType = iota
  CtrlChar
  // MouseEvent // Not implemented yet
)

// Non Printable Chars - Ctrl Char
const (
  CtrlAt int = iota
  CtrlA
  CtrlB
  CtrlC
  CtrlD
  CtrlE
  CtrlF
  CtrlG
  CtrlH // Backspace
  CtrlI
  CtrlJ // \n
  CtrlK
  CtrlL
  CtrlM
  CtrlN
  CtrlO
  CtrlP
  CtrlQ
  CtrlR
  CtrlS
  CtrlT
  CtrlU
  CtrlV
  CtrlW
  CtrlX
  CtrlY
  CtrlZ
  CtrlOpenBrkt // Ctrl + [ // ESC
  CtrlBckSlash // Ctrl + \
  CtrlCloseBrkt // Ctrl + ]
  CtrlCaret // Ctrl + ^
  CtrlUnderscore // Ctrl + _
)

// Non Printable Chars - Char
const (
  NUL int = iota // null
  SOH // start of heading 
  STX // start of text
  ETX // end of text 
  EOT // end of xmit
  ENQ // enquiry
  ACK // Acknowledge
  BEL // Bell
  BS // backspace
  HT // horizontal tab
  LF // line feed 
  VT // vertical tab
  FF // form feed
  CR // carriage feed
  SO // shift out
  SI // shift in
  DLE // data line escape
  DC1// device control 1
  DC2// device control 2
  DC3// device control 3
  DC4 // device control 4
  NAK // neg acknowledge
  SYN // sinchronous idel
  ETB // end of xmit block
  CAN // Cancel
  EM // End of Medium
  SUB // Substitute
  ESC // Escape
  FS // File seperator
  GS // Group separator
  RS // Record seperator
  US // Unit seperator
  DEL = 127 // Delete
)
