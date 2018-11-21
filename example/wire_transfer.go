package main

import (
	"fmt"
	"fsm"
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
				withdrawFrom := transfer.source
				transferTo := transfer.target
				amount := transfer.amount
				transfer.source <- amount
				return wt.Goto(AwaitFromState).With(
					&WireTransferData{withdrawFrom, transferTo, amount, wt},
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
