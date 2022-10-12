package src

import (
  "monkeyterm/src/core"
)


type InputEventBus struct {
  channel chan core.InputEventInstance
}

func NewInputEventBus() InputEventBus {
  return InputEventBus{
    channel: make(chan core.InputEventInstance),
  }
}

func (ieb InputEventBus) Send(event core.InputEventInstance) {
  ieb.channel <- event
}

func (ieb InputEventBus) Receive() chan core.InputEventInstance {
  return ieb.channel
}

func (ieb InputEventBus) Close() {
  close(ieb.channel)
}


type PollEventBus struct {
  channel chan core.PollEventInstance
}

func NewPollEventBus() PollEventBus {
  return PollEventBus{
    channel: make(chan core.PollEventInstance),
  }
}

func (peb PollEventBus) Send(event core.PollEventInstance) {
  peb.channel <- event
}

func (peb PollEventBus) Receive() chan core.PollEventInstance {
  return peb.channel
}

func (peb PollEventBus)  Close() {
  close(peb.channel)
}
