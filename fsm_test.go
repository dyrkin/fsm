package fsm

import (
	"fmt"
	"testing"
)

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
	coffeeMachine := NewFSM()
	coffeeMachine.StartWith(Open, newMachineData(0, 5, 1))

	coffeeMachine.When(Open)(
		func(event *Event) *NextState {
			machineData, machineDataOk := event.data.(*MachineData)
			deposit, depositOk := event.message.(*Deposit)
			setNumberOfCoffee, setNumberOfCoffeeOk := event.message.(*SetNumberOfCoffee)
			_, getNumberOfCoffeeOk := event.message.(*GetNumberOfCoffee)
			setCostOfCoffee, setCostOfCoffeeOk := event.message.(*SetCostOfCoffee)
			_, getCostOfCoffeeOk := event.message.(*GetCostOfCoffee)
			switch {
			case machineDataOk && machineData.coffeesLeft <= 0:
				return coffeeMachine.Goto(PoweredOff)

			case depositOk && machineDataOk && (deposit.value+machineData.currentTxTotal >= machineData.costOfCoffee):
				txTotal := machineData.currentTxTotal + deposit.value
				newData := newMachineData(txTotal, machineData.costOfCoffee, machineData.coffeesLeft)
				return coffeeMachine.Goto(ReadyToBuy).With(newData)

			case depositOk && machineDataOk && (deposit.value+machineData.currentTxTotal < machineData.costOfCoffee):
				txTotal := machineData.currentTxTotal + deposit.value
				newData := newMachineData(txTotal, machineData.costOfCoffee, machineData.coffeesLeft)
				return coffeeMachine.Stay().With(newData)

			case setNumberOfCoffeeOk && machineDataOk:
				fmt.Printf("Set new number of coffee: %d\n", setNumberOfCoffee.quantity)
				newData := newMachineData(machineData.currentTxTotal, machineData.costOfCoffee, setNumberOfCoffee.quantity)
				return coffeeMachine.Stay().With(newData)

			case getNumberOfCoffeeOk && machineDataOk:
				fmt.Printf("Coffees left: %d\n", machineData.coffeesLeft)
				return coffeeMachine.Stay()

			case setCostOfCoffeeOk && machineDataOk:
				fmt.Printf("Set new coffee price: %d\n", setCostOfCoffee.price)
				newData := newMachineData(machineData.currentTxTotal, setCostOfCoffee.price, machineData.coffeesLeft)
				return coffeeMachine.Stay().With(newData)

			case getCostOfCoffeeOk && machineDataOk:
				fmt.Printf("Cost of coffee: %d\n", machineData.costOfCoffee)
				return coffeeMachine.Stay()
			}
			return coffeeMachine.DefaultHandler()(event)
		})

	coffeeMachine.When(ReadyToBuy)(
		func(event *Event) *NextState {
			machineData, machineDataOk := event.data.(*MachineData)
			_, brewCoffeeOk := event.message.(*BrewCoffee)
			if brewCoffeeOk && machineDataOk {
				balanceToBeDispensed := machineData.currentTxTotal - machineData.costOfCoffee
				if balanceToBeDispensed > 0 {
					fmt.Printf("Balance to be dispensed is %d\n", balanceToBeDispensed)
					newData := newMachineData(0, machineData.costOfCoffee, machineData.coffeesLeft-1)
					return coffeeMachine.Goto(Open).With(newData)
				}
				newData := newMachineData(0, machineData.costOfCoffee, machineData.coffeesLeft-1)
				return coffeeMachine.Goto(Open).With(newData)
			}
			return coffeeMachine.DefaultHandler()(event)
		})

	coffeeMachine.When(PoweredOff)(
		func(event *Event) *NextState {
			_, startUpMachineOk := event.message.(*StartUpMachine)
			if startUpMachineOk {
				return coffeeMachine.Goto(Open)
			}
			fmt.Printf("Machine Powered down.  Please start machine first with StartUpMachine")
			return coffeeMachine.Stay()
		})

	coffeeMachine.SetDefaultHandler(
		func(event *Event) *NextState {
			_, shutDownMachineOk := event.message.(*ShutDownMachine)
			_, cancelOk := event.message.(*Cancel)
			machineData, machineDataOk := event.data.(*MachineData)

			switch {
			case shutDownMachineOk && machineDataOk:
				fmt.Printf("Balance is: %d\n", machineData.currentTxTotal)
				newData := newMachineData(0, machineData.costOfCoffee, machineData.coffeesLeft)
				return coffeeMachine.Goto(PoweredOff).With(newData)
			case cancelOk && machineDataOk:
				fmt.Printf("Balance is: %d\n", machineData.currentTxTotal)
				newData := newMachineData(0, machineData.costOfCoffee, machineData.coffeesLeft)
				return coffeeMachine.Goto(Open).With(newData)
			}
			panic("Something went wrong")
		})

	coffeeMachine.OnTransition(
		func(from State, to State) {
			switch {
			case from == Open && to == ReadyToBuy:
				fmt.Println("From Transacting to ReadyToBuy")
			case from == ReadyToBuy && to == Open:
				fmt.Println("From ReadyToBuy to Open")
			}
		})

	return coffeeMachine
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
}
