package util

import (
	"fmt"
	"time"
)

// Handler is the interface that provides the PopEvent function, which is invoked
// whenever an event is popped from the timer stack.
type Handler interface {
	PopEvent(eventID string, elapsed time.Duration)
}

// TimerStack provides a stack that allows for nested timing of named
// system events. Note that the timer not thread safe.
type TimerStack struct {
	stack           []*event
	popEventHandler Handler
}

type event struct {
	eventName string
	start     time.Time
}

// NewTimerStack creates a new timer stack.
func NewTimerStack(popEventHandler Handler) *TimerStack {
	return &TimerStack{stack: []*event{}, popEventHandler: popEventHandler}
}

type printHandler struct{}

// PopEvent prints a popped event ID and duration to std out.
func (p printHandler) PopEvent(eventID string, elapsed time.Duration) {
	fmt.Printf("[%s] %v\n", eventID, elapsed)
}

// NewPrintTimerStack create a new timer stack with basic output to standard out.
func NewPrintTimerStack() *TimerStack {
	return NewTimerStack(printHandler{})
}

// Push pushes an event onto the timer stack, saving its start time.
func (p *TimerStack) Push(eventName string) {
	p.stack = append(p.stack, &event{eventName, time.Now()})
}

// Pop pops the top of the timer stack and invokes the pop handler,
// passing it the popped event name and the time elapsed since the event
// was pushed.
func (p *TimerStack) Pop() {
	lastIndex := len(p.stack) - 1
	popped := p.stack[lastIndex]
	p.stack = p.stack[:lastIndex]
	elapsed := time.Since(popped.start)
	p.popEventHandler.PopEvent(popped.eventName, elapsed)
}
