package fsm

import (
	"fmt"
	"sync"
)

type State string

type Data interface{}

type Event struct {
	Message interface{}
	Data    Data
}

type NextState struct {
	state State
	data  Data
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
	defaultHandler     stateFunction
}

func NewFSM() *FSM {
	return &FSM{
		stateFunctions:     map[State]stateFunction{},
		mutex:              &sync.Mutex{},
		transitionFunction: func(from State, to State) {},
		defaultHandler: func(event *Event) *NextState {
			panic("Default handler is not defined")
		},
	}
}

func (fsm *FSM) StartWith(state State, data Data) {
	fsm.initialState = state
	fsm.initialData = data
	fsm.currentState = state
	fsm.currentData = data
}

func (fsm *FSM) Init() {
	fsm.StartWith(fsm.initialState, fsm.initialData)
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

func makeTransition(fsm *FSM, nextState *NextState) {
	fsm.transitionFunction(fsm.currentState, nextState.state)
	fsm.currentState = nextState.state
	fsm.currentData = nextState.data
}

func (fsm *FSM) Send(message interface{}) {
	mutex := fsm.mutex
	mutex.Lock()
	defer mutex.Unlock()
	currentState := fsm.currentState
	stateFunction := fsm.stateFunctions[currentState]
	nextState := stateFunction(&Event{message, fsm.currentData})
	makeTransition(fsm, nextState)
}

func (fsm *FSM) Goto(state State) *NextState {
	if _, ok := fsm.stateFunctions[state]; ok {
		return &NextState{state: state, data: fsm.currentData}
	}
	panic(fmt.Sprintf("Unknown state: %q", state))
}

func (fsm *FSM) Stay() *NextState {
	return &NextState{state: fsm.currentState, data: fsm.currentData}
}

func (ns *NextState) With(data Data) *NextState {
	ns.data = data
	return ns
}

func (fsm *FSM) DefaultHandler() stateFunction {
	return fsm.defaultHandler
}

func (fsm *FSM) OnTransition(f func(from State, to State)) {
	fsm.transitionFunction = f
}

func (fsm *FSM) CurrentState() State {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()
	return fsm.currentState
}

func (fsm *FSM) CurrentData() Data {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()
	return fsm.currentData
}
