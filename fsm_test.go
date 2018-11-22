package fsm

import (
	"fmt"
	"testing"
)

//the code is taken from https://github.com/arunma/AkkaFSM/blob/master/src/test/scala/me/rerun/akka/fsm/CoffeeSpec.scala

var Open State = "Open"
var ReadyToBuy State = "ReadyToBuy"
var PoweredOff State = "PoweredOff"

type MachineData struct {
	currentTxTotal int
	costOfCoffee   int
	coffeesLeft    int
}

type Deposit struct {
	value int
}

type Balance struct {
	value int
}

type Cancel struct{}
type BrewCoffee struct{}
type GetCostOfCoffee struct{}
type ShutDownMachine struct{}
type StartUpMachine struct{}
type SetNumberOfCoffee struct {
	quantity int
}
type SetCostOfCoffee struct {
	price int
}
type GetNumberOfCoffee struct{}

func newMachineData(currentTxTotal int, costOfCoffee int, coffeesLeft int) *MachineData {
	return &MachineData{currentTxTotal: currentTxTotal, costOfCoffee: costOfCoffee, coffeesLeft: coffeesLeft}
}

func newCoffeeMachine() *FSM {
	cm := NewFSM()
	cm.StartWith(Open, newMachineData(0, 5, 1))

	cm.When(Open)(
		func(event *Event) *NextState {
			machineData, machineDataOk := event.Data.(*MachineData)
			deposit, depositOk := event.Message.(*Deposit)
			setNumberOfCoffee, setNumberOfCoffeeOk := event.Message.(*SetNumberOfCoffee)
			_, getNumberOfCoffeeOk := event.Message.(*GetNumberOfCoffee)
			setCostOfCoffee, setCostOfCoffeeOk := event.Message.(*SetCostOfCoffee)
			_, getCostOfCoffeeOk := event.Message.(*GetCostOfCoffee)
			switch {
			case machineDataOk && machineData.coffeesLeft <= 0:
				return cm.Goto(PoweredOff)

			case depositOk && machineDataOk && (deposit.value+machineData.currentTxTotal >= machineData.costOfCoffee):
				txTotal := machineData.currentTxTotal + deposit.value
				newData := newMachineData(txTotal, machineData.costOfCoffee, machineData.coffeesLeft)
				return cm.Goto(ReadyToBuy).With(newData)

			case depositOk && machineDataOk && (deposit.value+machineData.currentTxTotal < machineData.costOfCoffee):
				txTotal := machineData.currentTxTotal + deposit.value
				newData := newMachineData(txTotal, machineData.costOfCoffee, machineData.coffeesLeft)
				return cm.Stay().With(newData)

			case setNumberOfCoffeeOk && machineDataOk:
				fmt.Printf("Set new number of coffee: %d\n", setNumberOfCoffee.quantity)
				newData := newMachineData(machineData.currentTxTotal, machineData.costOfCoffee, setNumberOfCoffee.quantity)
				return cm.Stay().With(newData)

			case getNumberOfCoffeeOk && machineDataOk:
				fmt.Printf("Coffees left: %d\n", machineData.coffeesLeft)
				return cm.Stay()

			case setCostOfCoffeeOk && machineDataOk:
				fmt.Printf("Set new coffee price: %d\n", setCostOfCoffee.price)
				newData := newMachineData(machineData.currentTxTotal, setCostOfCoffee.price, machineData.coffeesLeft)
				return cm.Stay().With(newData)

			case getCostOfCoffeeOk && machineDataOk:
				fmt.Printf("Cost of coffee: %d\n", machineData.costOfCoffee)
				return cm.Stay()
			}
			return cm.DefaultHandler()(event)
		})

	cm.When(ReadyToBuy)(
		func(event *Event) *NextState {
			machineData, machineDataOk := event.Data.(*MachineData)
			_, brewCoffeeOk := event.Message.(*BrewCoffee)
			if brewCoffeeOk && machineDataOk {
				balanceToBeDispensed := machineData.currentTxTotal - machineData.costOfCoffee
				if balanceToBeDispensed > 0 {
					fmt.Printf("Balance to be dispensed is %d\n", balanceToBeDispensed)
					newData := newMachineData(0, machineData.costOfCoffee, machineData.coffeesLeft-1)
					return cm.Goto(Open).With(newData)
				}
				newData := newMachineData(0, machineData.costOfCoffee, machineData.coffeesLeft-1)
				return cm.Goto(Open).With(newData)
			}
			return cm.DefaultHandler()(event)
		})

	cm.When(PoweredOff)(
		func(event *Event) *NextState {
			_, startUpMachineOk := event.Message.(*StartUpMachine)
			if startUpMachineOk {
				return cm.Goto(Open)
			}
			fmt.Printf("Machine Powered down.  Please start machine first with StartUpMachine")
			return cm.Stay()
		})

	cm.SetDefaultHandler(
		func(event *Event) *NextState {
			_, shutDownMachineOk := event.Message.(*ShutDownMachine)
			_, cancelOk := event.Message.(*Cancel)
			machineData, machineDataOk := event.Data.(*MachineData)

			switch {
			case shutDownMachineOk && machineDataOk:
				fmt.Printf("Balance is: %d\n", machineData.currentTxTotal)
				newData := newMachineData(0, machineData.costOfCoffee, machineData.coffeesLeft)
				return cm.Goto(PoweredOff).With(newData)
			case cancelOk && machineDataOk:
				fmt.Printf("Balance is: %d\n", machineData.currentTxTotal)
				newData := newMachineData(0, machineData.costOfCoffee, machineData.coffeesLeft)
				return cm.Goto(Open).With(newData)
			}
			panic("Something went wrong")
		})

	cm.OnTransition(
		func(from State, to State) {
			switch {
			case from == Open && to == ReadyToBuy:
				fmt.Println("From Transacting to ReadyToBuy")
			case from == ReadyToBuy && to == Open:
				fmt.Println("From ReadyToBuy to Open")
			}
		})

	return cm
}

func TestCoffeeMachine(t *testing.T) {
	//should allow setting and getting of price of coffee
	coffeeMachine := newCoffeeMachine()
	coffeeMachine.Send(&SetCostOfCoffee{7})
	coffeeMachine.Send(&GetCostOfCoffee{})

	machineData := coffeeMachine.CurrentData().(*MachineData)

	if machineData.costOfCoffee != 7 {
		t.Errorf("Cost of coffee is not 7. Price: %d", machineData.costOfCoffee)
	}

	//should allow setting and getting of maximum number of coffees
	coffeeMachine = newCoffeeMachine()
	coffeeMachine.Send(&SetNumberOfCoffee{10})
	coffeeMachine.Send(&GetNumberOfCoffee{})

	machineData = coffeeMachine.CurrentData().(*MachineData)

	if machineData.coffeesLeft != 10 {
		t.Errorf("Number of coffee is not 10. Number: %d", machineData.coffeesLeft)
	}

	//should stay at Transacting when the Deposit is less then the price of the coffee
	coffeeMachine = newCoffeeMachine()
	coffeeMachine.Send(&SetCostOfCoffee{5})
	coffeeMachine.Send(&SetNumberOfCoffee{10})

	if coffeeMachine.CurrentState() != Open {
		t.Errorf("Current state is not Open. State: %s", coffeeMachine.CurrentState())
	}

	coffeeMachine.Send(&Deposit{2})
	coffeeMachine.Send(&GetNumberOfCoffee{})

	machineData = coffeeMachine.CurrentData().(*MachineData)
	if machineData.coffeesLeft != 10 {
		t.Errorf("Number of coffee is not 10. Number: %d", machineData.coffeesLeft)
	}

	if machineData.currentTxTotal != 2 {
		t.Errorf("Deposit is not 2. Deposit: %d", machineData.currentTxTotal)
	}

	//should transition to ReadyToBuy and then Open when the Deposit is equal to the price of the coffee
	coffeeMachine = newCoffeeMachine()
	coffeeMachine.Send(&SetCostOfCoffee{5})
	coffeeMachine.Send(&SetNumberOfCoffee{10})

	if coffeeMachine.CurrentState() != Open {
		t.Errorf("Current state is not Open. State: %s", coffeeMachine.CurrentState())
	}

	coffeeMachine.Send(&Deposit{5})
	coffeeMachine.Send(&BrewCoffee{})
	coffeeMachine.Send(&GetNumberOfCoffee{})

	machineData = coffeeMachine.CurrentData().(*MachineData)

	if machineData.coffeesLeft != 9 {
		t.Errorf("Number of coffee is not 9. Number: %d", machineData.coffeesLeft)
	}

	//should transition to Open after flushing out all the deposit when the coffee is canceled
	coffeeMachine = newCoffeeMachine()
	coffeeMachine.Send(&SetCostOfCoffee{5})
	coffeeMachine.Send(&SetNumberOfCoffee{10})

	if coffeeMachine.CurrentState() != Open {
		t.Errorf("Current state is not Open. State: %s", coffeeMachine.CurrentState())
	}

	coffeeMachine.Send(&Deposit{2})
	coffeeMachine.Send(&Deposit{2})
	coffeeMachine.Send(&Deposit{2})

	machineData = coffeeMachine.CurrentData().(*MachineData)

	if machineData.currentTxTotal != 6 {
		t.Errorf("Deposit is not 6. Deposit: %d", machineData.currentTxTotal)
	}

	coffeeMachine.Send(&Cancel{})

	machineData = coffeeMachine.CurrentData().(*MachineData)

	if machineData.currentTxTotal != 0 {
		t.Errorf("Deposit is not 0. Deposit: %d", machineData.currentTxTotal)
	}

	//should transition to PoweredOff state if the machine is shut down from ReadyToBuyState
	coffeeMachine = newCoffeeMachine()
	coffeeMachine.Send(&SetCostOfCoffee{5})
	coffeeMachine.Send(&SetNumberOfCoffee{10})

	if coffeeMachine.CurrentState() != Open {
		t.Errorf("Current state is not Open. State: %s", coffeeMachine.CurrentState())
	}

	coffeeMachine.Send(&Deposit{2})
	coffeeMachine.Send(&Deposit{2})
	coffeeMachine.Send(&Deposit{2})

	if coffeeMachine.CurrentState() != ReadyToBuy {
		t.Errorf("Current state is not ReadyToBuy. State: %s", coffeeMachine.CurrentState())
	}

	coffeeMachine.Send(&ShutDownMachine{})

	if coffeeMachine.CurrentState() != PoweredOff {
		t.Errorf("Current state is not PoweredOff. State: %s", coffeeMachine.CurrentState())
	}

	machineData = coffeeMachine.CurrentData().(*MachineData)

	if machineData.currentTxTotal != 0 {
		t.Errorf("Deposit is not 0. Deposit: %d", machineData.currentTxTotal)
	}

	//should open the machine to operation if powered on
	coffeeMachine = newCoffeeMachine()
	coffeeMachine.Send(&SetCostOfCoffee{5})
	coffeeMachine.Send(&SetNumberOfCoffee{10})

	if coffeeMachine.CurrentState() != Open {
		t.Errorf("Current state is not Open. State: %s", coffeeMachine.CurrentState())
	}

	coffeeMachine.Send(&Deposit{2})
	coffeeMachine.Send(&Deposit{2})

	coffeeMachine.Send(&ShutDownMachine{})

	if coffeeMachine.CurrentState() != PoweredOff {
		t.Errorf("Current state is not PoweredOff. State: %s", coffeeMachine.CurrentState())
	}
}
