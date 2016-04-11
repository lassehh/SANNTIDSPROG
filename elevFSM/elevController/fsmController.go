package elevController

import (
	. "./elevDrivers"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	IDLE       = 0
	DRIVING    = 1
	DOOR_TIMER = 2
)

type Elevator struct {
	STATE             int
	CURRENT_FLOOR     int
	DESTINATION_FLOOR int
	DIRECTION         int
}

func FSM_setup_elevator() {
	Elev_init()
	Elev_set_motor_direction(DIRN_DOWN)
	for {
		if Elev_get_floor_sensor_signal() != -1 {
			Elev_set_motor_direction(DIRN_STOP)
			break
		}
	}
	Elev_set_floor_indicator(Elev_get_floor_sensor_signal())
}

func FSM_create_elevator() Elevator {
	e := Elevator{STATE: IDLE, CURRENT_FLOOR: Elev_get_floor_sensor_signal(), DESTINATION_FLOOR: Elev_get_floor_sensor_signal(), DIRECTION: DIRN_STOP}
	return e
}

func FSM_Start_Driving(NewObjective Button, e *Elevator, State_Chan chan int, Motor_Direction_Chan chan int, Location_Chan chan int) {
	if e.CURRENT_FLOOR > NewObjective.Floor {
		Elev_set_motor_direction(-1)
		Motor_Direction_Chan <- -1
		fmt.Printf("\nThe elevator started driving in %d direction", -1)
		State_Chan <- DRIVING
		fmt.Println("\nState message sent")
	}
	if e.CURRENT_FLOOR < NewObjective.Floor {
		Elev_set_motor_direction(1)
		Motor_Direction_Chan <- 1
		fmt.Printf("\nThe elevator started driving in %d direction", 1)
		State_Chan <- DRIVING
		fmt.Println("\nState message sent")
	}
	if e.CURRENT_FLOOR == NewObjective.Floor {
		State_Chan <- DRIVING
	}
}

func FSM_objective_dealer(e *Elevator, State_Chan chan int, Destination_Chan chan int, Objective_Chan chan Button) {
	for {
		time.Sleep(time.Millisecond * 200)
		nextOrder := Next_order()
		//fmt.Println(nextOrder.Floor)
		if e.STATE == IDLE && nextOrder.Floor != -1 {
			Objective_Chan <- nextOrder
			Destination_Chan <- nextOrder.Floor
			fmt.Println("\n\nA new objective was sent to the elevator")
		}
	}
}

func FSM_elevator_updater(e *Elevator, Motor_Direction_Chan chan int, Location_Chan chan int, Destination_Chan chan int, State_Chan chan int) {
	for{
		select {
		case NewDirection := <-Motor_Direction_Chan:
			e.DIRECTION = NewDirection
			fmt.Println("\nNew direction: ", NewDirection)
		case NewFloor := <-Location_Chan:
			e.CURRENT_FLOOR = NewFloor
			fmt.Println("\nNew location: ", NewFloor)
		case NewDestination := <-Destination_Chan:
			e.DESTINATION_FLOOR = NewDestination
			fmt.Println("\nNew destination: ", NewDestination)
		case NewState := <-State_Chan:
			e.STATE = NewState
			fmt.Println("\nNew state: ", NewState)
		}
	}
}

func FSM_floor_tracker(e *Elevator, Location_Chan chan int, Floor_Arrival_Chan chan int) {
	for {
		time.Sleep(time.Millisecond * 200)
		if Elev_get_floor_sensor_signal() != -1 {
			NewFloor := Elev_get_floor_sensor_signal()
			Elev_set_floor_indicator(NewFloor)
			Location_Chan <- NewFloor
			Floor_Arrival_Chan <- NewFloor
			fmt.Printf("\nThe elevator arrived at the %d. floor", NewFloor)
		}
	}
}

func FSM_sensor_pooler(Button_Press_Chan chan Button) {
	for {
		for button := B_UP; button <= B_COMMAND; button++ {
			for floor := 0; floor < N_FLOORS; floor++ {
				if button == B_UP && floor == N_FLOORS-1 {
					continue
				}
				if button == B_DOWN && floor == 0 {
					continue
				}
				button_signal := Elev_get_button_signal(button, floor)
				if button_signal == 1 {
					Button_Press_Chan <- Button{Button_type: button, Floor: floor}
					fmt.Println("\nA button was pressed...")
				}
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func FSM_should_stop_or_not(newFloorArrival int, e *Elevator, State_Chan chan int, Motor_Direction_Chan chan int, Door_Open_Req_Chan chan int) {
	if newFloorArrival == e.DESTINATION_FLOOR && e.STATE == DRIVING {
		Elev_set_motor_direction(0)
		Motor_Direction_Chan <- 0
		Door_Open_Req_Chan <- 1
		State_Chan <- DOOR_TIMER
		fmt.Println("\nThe elevator decided to stop")
	}
	fmt.Println("\nThe elevator decided not to stop")
}

func FSM_door_opener(doorReq int, e *Elevator, State_Chan chan int) {
	Elev_set_door_open_lamp(1)
	fmt.Println("\nThe door has been opened")

	time.Sleep(time.Second * 2)

	Elev_set_door_open_lamp(0)

	State_Chan <- IDLE
	Remove_order(e.CURRENT_FLOOR)
}

func FSM_safekill() {
	var c = make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	Elev_set_motor_direction(0)
	log.Printf("\nUser terminated program")
	os.Exit(1)
}