package main

import (
	. "./elevController"
	"fmt"
	"time"
)

func main() {
	/* INITIALIZATION */
	FSM_setup_elevator()

	/* SETS INITIAL STATE VARIABLES */
	e := 							FSM_create_elevator()

	/* CHANNELS FOR UPDATING THE ELEVATOR VARIABLES */
	Button_Press_Chan := 			make(chan Button, 10)
	Location_Chan := 				make(chan int, 1)
	Motor_Direction_Chan := 		make(chan int, 1)
	Destination_Chan := 			make(chan int, 1)
	State_Chan := 					make(chan int, 1)

	/* EVENT CHANNELS */
	Objective_Chan := 				make(chan Button, 1)
	Floor_Arrival_Chan := 			make(chan int, 1)
	Door_Open_Req_Chan := 			make(chan int, 1)

	/* STARTS ESSENTIAL PROCESSES */
	Orders_init()
	go Order_handler(Button_Press_Chan)
	go FSM_safekill()
	go FSM_sensor_pooler(Button_Press_Chan)
	go FSM_floor_tracker(&e, Location_Chan, Floor_Arrival_Chan)
	go FSM_objective_dealer(&e, State_Chan, Destination_Chan, Objective_Chan)
	go FSM_elevator_updater(&e, Motor_Direction_Chan, Location_Chan, Destination_Chan, State_Chan)
	time.Sleep(time.Millisecond * 200)

	/* STARTUP TEXT */
	fmt.Printf("\n\n\n####################################################\n")
	fmt.Printf("## The elevator has been succesfully initiated! #### \n")
	fmt.Printf("####################################################\n\n")
	fmt.Printf("STATE: %d , ", e.STATE)
	fmt.Printf("CURRENT_FLOOR: %d , ", e.CURRENT_FLOOR)
	fmt.Printf("DESTINATION_FLOOR: %d , ", e.DESTINATION_FLOOR)
	fmt.Printf("DIRECTION: %d \n\n\n", e.DIRECTION)

	Print_all_orders()

	for {
		select {
		case newObjective := 		<-Objective_Chan:
			FSM_Start_Driving(newObjective, &e, State_Chan, Motor_Direction_Chan, Location_Chan)
		
		case newFloorArrival := 	<-Floor_Arrival_Chan:
			FSM_should_stop_or_not(newFloorArrival, &e, State_Chan, Motor_Direction_Chan, Door_Open_Req_Chan)
		
		case doorReq := 			<-Door_Open_Req_Chan:
			FSM_door_opener(doorReq, &e, State_Chan)
		}
	}
}
