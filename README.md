# Finite State Machine for Go

## Overview

The FSM (Finite State Machine) is best described in the [Erlang design principles](http://www.erlang.org/documentation/doc-4.8.2/doc/design_principles/fsm.html)

A FSM can be described as a set of relations of the form:
> State(S) x Event(E) -> Actions (A), State(S')

These relations are interpreted as meaning:
> If we are in state S and the event E occurs, we should perform the actions A and make a transition to the state S'.

## A Simple Example

```go
import (
	"fmt"
	"github.com/dyrkin/fsm"
)

//states
const InitialState = "Initial"
const AwaitFromState = "AwaitFrom"
const AwaitToState = "AwaitTo"
const DoneState = "Done"

//messages
type Transfer struct {
	source chan int
	target chan int
	amount int
}

const Done = "Done"
const Failed = "Failed"

//data
type WireTransferData struct {
	source chan int
	target chan int
	amount int
	client *fsm.FSM
}

func newWireTransfer(transferred chan bool) *fsm.FSM {
	wt := fsm.NewFSM()

	wt.StartWith(InitialState, nil)

	wt.When(InitialState)(
		func(event *fsm.Event) *fsm.NextState {
			transfer, transferOk := event.Message.(*Transfer)
			if transferOk && event.Data == nil {
				transfer.source <- transfer.amount
				return wt.Goto(AwaitFromState).With(
					&WireTransferData{transfer.source, transfer.target, transfer.amount, wt},
				)
			}
			return wt.DefaultHandler()(event)
		})

	wt.When(AwaitFromState)(
		func(event *fsm.Event) *fsm.NextState {
			data, dataOk := event.Data.(*WireTransferData)
			if dataOk {
				switch event.Message {
				case Done:
					data.target <- data.amount
					return wt.Goto(AwaitToState)
				case Failed:
					go data.client.Send(Failed)
					return wt.Stay()
				}
			}
			return wt.DefaultHandler()(event)
		})

	wt.When(AwaitToState)(
		func(event *fsm.Event) *fsm.NextState {
			data, dataOk := event.Data.(*WireTransferData)
			if dataOk {
				switch event.Message {
				case Done:
					transferred <- true
					return wt.Stay()
				case Failed:
					go data.client.Stay()
				}
			}
			return wt.DefaultHandler()(event)
		})
	return wt
}
```

The code is pretty self explanatory. The state machine will start in the Initial state with all values uninitialized. The only type of message which can be received in the Initial state is the initial Transfer request at which point a withdraw amount is sent to the source account and the state machine transitions to the AwaitFrom state.

When the system is in the AwaitFrom state the only two messages that can be received are Done or Failure from the source account. If the Done business acknowledgement is received the system will send a deposit amount to the target account and transition to the AwaitTo state.

When the system is in the AwaitTo state the only two messages that can be received are the Done or Failure from the target account.

To run the code above you can use the following code:

```go
func main() {

	transferred := make(chan bool)

	wireTransfer := newWireTransfer(transferred)

	transfer := &Transfer{
		source: make(chan int),
		target: make(chan int),
		amount: 30,
	}

	source := func() {
		withdrawAmount := <-transfer.source
		fmt.Printf("Withdrawn from source account: %d\n", withdrawAmount)
		wireTransfer.Send(Done)
	}

	target := func() {
		topupAmount := <-transfer.target
		fmt.Printf("ToppedUp target account: %d\n", topupAmount)
		wireTransfer.Send(Done)
	}

	go source()
	go target()

	go wireTransfer.Send(transfer)

	if done := <-transferred; !done {
		panic("Something went wrong")
	}

	fmt.Println("DONE")
}
```

It will produce the following output:

> Withdrawn from source account: 30  
ToppedUp target account: 30  
DONE