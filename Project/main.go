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
	Orders_init()
	e := FSM_create_elevator()
	e_system := Initialize_elev_system()

	/* CHANNELS FOR UPDATING THE ELEVATOR VARIABLES */
	Button_Press_Chan := make(chan Button, 10)
	Location_Chan := make(chan int, 1)
	Motor_Direction_Chan := make(chan int, 1)
	Destination_Chan := make(chan int, 1)
	State_Chan := make(chan int, 1)

	/* EVENT CHANNELS */
	Objective_Chan := make(chan Button, 1)
	Floor_Arrival_Chan := make(chan int, 1)
	Door_Open_Req_Chan := make(chan int, 1)

	/* MESSAGE CHANNELS */
	Rchv_message_Chan := make(chan Message)
	Broadcast_message_Chan := make(chan Message)
	//Master_Ready_To_Send_Chan := make(chan bool)
	Broadcast_Elev_System_Chan := make(chan Elevator_System)
	Rchv_Elev_System_Chan := make(chan Elevator_System)
	To_Master_Chan := make(chan Message)

	Master_Send_Timer_Chan := make(chan bool, 1)
	Master_Req_Update_Chan := make(chan bool)

	/* STARTS ESSENTIAL PROCESSES */
	go Order_handler(Button_Press_Chan)
	go Get_internal_orders(&e)
	go FSM_safekill()
	go FSM_sensor_pooler(Button_Press_Chan)
	go FSM_floor_tracker(&e, Location_Chan, Floor_Arrival_Chan)
	go FSM_objective_dealer(&e, State_Chan, Destination_Chan, Objective_Chan)
	go FSM_elevator_updater(&e, Motor_Direction_Chan, Location_Chan, Destination_Chan, State_Chan)

	/* STARTS THE NETWORK BETWEEN THE ELEVATORS AND THE MESSAGE-PASSING */
	go MessageSetter(Broadcast_message_Chan, &e_system, &e)
	//go UDPSend(PORT, Broadcast_message_Chan, From_Master_Chan,Broadcast_Elev_System)
	go UDPListen(Is_elev_master(e_system), PORT, Get_Master_IP(e_system), Rchv_message_Chan)

	time.Sleep(time.Millisecond * 200)

	//Master_Send_Timer_Chan <- true //just for the test of it. As far as we have implemented, there needs to be a slave for Master_Send_Timer_Chan to work
	/* STARTUP TEXT */
	fmt.Printf("\n\n\n####################################################\n")
	fmt.Printf("## The elevator has been succesfully initiated! #### \n")
	fmt.Printf("####################################################\n\n")
	fmt.Printf("STATE: %d , ", e.State)
	fmt.Printf("CURRENT_FLOOR: %d , ", e.CurrentFloor)
	fmt.Printf("DESTINATION_FLOOR: %d , ", e.DestinationFloor)
	fmt.Printf("DIRECTION: %d \n\n\n", e.Direction)
	//msg := CreateMessage()
	//fmt.Println(msg)
	//Rchv_message_Chan <- msg

	Print_all_orders()

	for {
		select {
		case newObjective := <-Objective_Chan:
			FSM_Start_Driving(newObjective, &e, State_Chan, Motor_Direction_Chan, Location_Chan)

		case newFloorArrival := <-Floor_Arrival_Chan:
			FSM_should_stop_or_not(newFloorArrival, &e, State_Chan, Motor_Direction_Chan, Door_Open_Req_Chan)

		case doorReq := <-Door_Open_Req_Chan:
			FSM_door_opener(doorReq, &e, State_Chan)

		case msgToMaster := <-To_Master_Chan: //not needed?
			Message_Compiler_Master(msgToMaster, &e_system)

		case elevSystemToSlave := <-Rchv_Elev_System_Chan:
			Sync_with_system(elevSystemToSlave, &e, &e_system)

		case slaveSendUpdate := <-Master_Req_Update_Chan:
			fmt.Println(slaveSendUpdate)
			UDPSendToMaster(PORT, Broadcast_message_Chan)
			go Timer(Master_Send_Timer_Chan)

		case masterSendUpdate := <-Master_Send_Timer_Chan:
			fmt.Println(masterSendUpdate)
			rchvMessage := <-Rchv_message_Chan
			Message_Compiler_Master(rchvMessage, &e_system)
			UDPSendToSlave(PORT, Broadcast_Elev_System_Chan) //needs to implement such that master sendt to itself
			go Timer(Master_Req_Update_Chan)
		}

	}
}
