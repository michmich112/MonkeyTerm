package core

import "sync"


type EventType int
type Events map[EventType] interface{}

type EventBus struct {
  events Events
  cond *sync.Cond
  ignore map[EventType] bool
}


func NewEventBus() *EventBus {
  return &EventBus{
    events: make(Events),
    cond: sync.NewCond(&sync.Mutex{}),
    ignore: make(map[EventType] bool),
  }
}

func (eb *EventBus) Wait(callback func(*Events)) {
  eb.cond.L.Lock()
  if len(eb.events) == 0 {
    eb.cond.Wait()
  }

  callback(&eb.events)
  eb.cond.L.Unlock()
}
