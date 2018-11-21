package fsm

import (
	"fmt"
	"sync"
)

type Event struct {
	message interface{}
	data    interface{}
}

type State string
type Data interface{}

type NextState struct {
	state Option
	data  Data
}

type None struct{}

type Some struct {
	state State
}

type Option interface {
	Value() State
}

func (o *Some) Value() State {
	return o.state
}

func (o *None) Value() State {
	return "end"
}

type stateFunction func(event *Event) *NextState

type FSM struct {
	initialState       State
	initialData        Data
	currentState       State
	currentData        Data
	stateFunctions     map[State]stateFunction
	transitionFunction func(from State, to State)
	mutex              *sync.Mutex
	completed          bool
	asyncEventQueue    chan *Event
	defaultHandler     stateFunction
}

func NewFSM() *FSM {
	return &FSM{
		stateFunctions:     map[State]stateFunction{},
		mutex:              &sync.Mutex{},
		transitionFunction: func(from State, to State) {},
		asyncEventQueue:    make(chan *Event),
		defaultHandler: func(event *Event) *NextState {
			panic("Default handler is not defined")
		},
	}
}

func (fsm *FSM) When(state State) func(stateFunction) *FSM {
	return func(f stateFunction) *FSM {
		fsm.stateFunctions[state] = f
		return fsm
	}
}

func (fsm *FSM) SetDefaultHandler(defaultHandler stateFunction) {
	fsm.defaultHandler = defaultHandler
}

func (fsm *FSM) StartWith(state State, data Data) {
	fsm.completed = false
	fsm.initialState = state
	fsm.initialData = data
	fsm.currentState = state
	fsm.currentData = data
}

func (fsm *FSM) Send(message interface{}) {
	mutex := fsm.mutex
	mutex.Lock()
	defer mutex.Unlock()
	if fsm.completed {
		panic("FSM reached its final state. Call Init() to reinitialize FSM")
	}
	currentState := fsm.currentState
	stateFunction := fsm.stateFunctions[currentState]
	nextState := stateFunction(&Event{message, fsm.currentData})
	fsm.makeTransition(nextState)
}

func (fsm *FSM) makeTransition(nextState *NextState) {
	fsm.transitionFunction(fsm.currentState, nextState.state.Value())
	switch s := nextState.state.(type) {
	case *Some:
		fmt.Printf("Transition from %q to %q\n", fsm.currentState, s.Value())
		fsm.currentState = s.Value()
	default:
		fmt.Printf("Transition from %q to %q\n", fsm.currentState, s.Value())
		fsm.completed = true
	}
	fsm.currentData = nextState.data
}

func (fsm *FSM) Goto(state State) *NextState {
	return &NextState{state: &Some{state}, data: fsm.currentData}
}

func (fsm *FSM) Stay() *NextState {
	return &NextState{state: &Some{fsm.currentState}, data: fsm.currentData}
}

func (fsm *FSM) End() *NextState {
	return &NextState{state: &None{}, data: fsm.currentData}
}

func (fsm *FSM) DefaultHandler() stateFunction {
	return fsm.defaultHandler
}

func (fsm *FSM) Init() {
	fsm.StartWith(fsm.initialState, fsm.initialData)
}

func (ns *NextState) With(data Data) *NextState {
	ns.data = data
	return ns
}

func (fsm *FSM) OnTransition(f func(from State, to State)) {
	fsm.transitionFunction = f
}

func (fsm *FSM) CurrentState() State {
	return fsm.currentState
}

func (fsm *FSM) CurrentData() Data {
	return fsm.currentData
}
