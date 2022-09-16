package core

import "time"

//
// type EventType int
// type Events map[EventType] interface{}
//
// type EventBus struct {
//   events Events
//   cond *sync.Cond
//   ignore map[EventType] bool
// }
//
//
// func NewEventBus() *EventBus {
//   return &EventBus{
//     events: make(Events),
//     cond: sync.NewCond(&sync.Mutex{}),
//     ignore: make(map[EventType] bool),
//   }
// }
//
// func (eb *EventBus) Wait(callback func(*Events)) {
//   eb.cond.L.Lock()
//   if len(eb.events) == 0 {
//     eb.cond.Wait()
//   }
//
//   callback(&eb.events)
//   eb.cond.L.Unlock()
// }

type EventBus[D any] interface {
  // Send a message through the event bus
  Send(Event[D])

  // Receive Events through chan 
  Receive() chan Event[D]

  // Closes the event bus and the channeld:w
  Close()
}


type Event[D any] interface {
  GetTimestamp() (time.Time)
  GetEventType() (int)
  GetEventTrigger() (int)
  GetEventValue() D // interface{} // this value will need to be cast depending on the type of the event
}

// could use generrics]


type EventTrigger = int

// Event triggers
const (
  Poll EventTrigger = iota // Timer poll = refresh rate
  Input // User Inputs a character
  System // System triggers a system call
  Resize // User resizes window
  Callback // Used for custom events
)

type InputEvent struct {
  Type InputEventType
  T time.Time
  Value int
}

type InputEventType = int

const (
  InputEvent_Rune InputEventType = iota
  InputEvent_CtrlChar
  InputEvent_EscSequence
  InputEvent_ReadError
)

type InputEventValueType = int

func (ie InputEvent) GetTimestamp() time.Time {return ie.T}
func (ie InputEvent) GetEventType() int {return ie.Type}
func (ie InputEvent) GetEventTrigger() int {return System}
func (ie InputEvent) GetEventValue() InputEventValueType {return ie.Value}

type InputEventHandler = func(event Event[InputEventValueType])

type InputEventInstance = Event[InputEventValueType]


type PollEvent struct {
  T time.Time
}

type PollEventValueType = time.Time

func (pe PollEvent) GetTimestamp() time.Time {return pe.T}
func (pe PollEvent) GetEventType() int {return 0}
func (pe PollEvent) GetEventTrigger() int {return Poll}
func (pe PollEvent) GetEventValue() time.Time {return pe.T}

type PollEventHandler = func(event Event[PollEventValueType])

type PollEventInstance = Event[PollEventValueType]
